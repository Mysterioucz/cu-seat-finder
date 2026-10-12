// Package broker is the RabbitMQ plumbing for async events between services
// (ADR-007). Delivery is at-least-once: publishers wait for a broker confirm,
// and consumers must be idempotent because a message can arrive twice.
//
// The whole topology (exchange, queues, dead-letter queue, bindings) is
// declared here and by both sides, so it does not matter which service boots
// first: a message published before the consumer ever ran still lands in its
// durable queue instead of being dropped as unroutable.
package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// Exchange carries every reservation lifecycle event; the routing key is the event name.
	Exchange = "reservation.events"

	RoutingNoShow = "reservation.no_show"

	// QueueUserNoShow is User service's inbox for no-show events (penalties).
	QueueUserNoShow = "user.reservation-no-show"
	// queueDeadLetter holds messages a consumer rejected as unprocessable.
	queueDeadLetter = "reservation.events.dlq"
)

// NoShowEvent is published by Reservation when an accepted booking passes its
// check-in deadline without a check-in. ReservationID doubles as the
// idempotency key on the consumer side.
type NoShowEvent struct {
	ReservationID   string    `json:"reservation_id"`
	UserID          string    `json:"user_id"`
	RoomID          string    `json:"room_id"`
	StartTime       time.Time `json:"start_time"`
	CheckinDeadline time.Time `json:"checkin_deadline"`
}

func declareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(Exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(queueDeadLetter, true, false, false, false, nil); err != nil {
		return err
	}
	// Dead-letter through the default exchange, which routes by queue name.
	args := amqp.Table{"x-dead-letter-exchange": "", "x-dead-letter-routing-key": queueDeadLetter}
	if _, err := ch.QueueDeclare(QueueUserNoShow, true, false, false, false, args); err != nil {
		return err
	}
	return ch.QueueBind(QueueUserNoShow, RoutingNoShow, Exchange, false, nil)
}

// Publisher keeps one confirm-mode channel and reconnects lazily: if RabbitMQ
// is down, Publish returns an error and the caller retries on its next tick.
type Publisher struct {
	url  string
	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(url string) *Publisher { return &Publisher{url: url} }

func (p *Publisher) channel() (*amqp.Channel, error) {
	if p.ch != nil && !p.ch.IsClosed() {
		return p.ch, nil
	}
	p.closeLocked()
	conn, err := amqp.Dial(p.url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err == nil {
		err = declareTopology(ch)
	}
	if err == nil {
		err = ch.Confirm(false)
	}
	if err != nil {
		conn.Close()
		return nil, err
	}
	p.conn, p.ch = conn, ch
	return ch, nil
}

// Publish sends a persistent JSON message and blocks until RabbitMQ confirms
// it has taken responsibility for it. Only a nil return means "delivered".
func (p *Publisher) Publish(ctx context.Context, routingKey, messageID string, event any) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	ch, err := p.channel()
	if err != nil {
		return fmt.Errorf("broker connect: %w", err)
	}
	dc, err := ch.PublishWithDeferredConfirmWithContext(ctx, Exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    messageID,
		Timestamp:    time.Now(),
		Type:         routingKey,
		Body:         body,
	})
	if err != nil {
		p.closeLocked()
		return fmt.Errorf("broker publish: %w", err)
	}
	acked, err := dc.WaitContext(ctx)
	if err != nil {
		p.closeLocked()
		return fmt.Errorf("broker confirm: %w", err)
	}
	if !acked {
		return errors.New("broker publish: nacked by broker")
	}
	return nil
}

func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closeLocked()
}

func (p *Publisher) closeLocked() {
	if p.conn != nil {
		p.conn.Close()
	}
	p.conn, p.ch = nil, nil
}

// ErrPermanent marks a handler error as one retrying will never fix (malformed
// body, unknown user): wrap it with fmt.Errorf("%w: ...", ErrPermanent) and the
// message goes to the dead-letter queue instead of being redelivered forever.
var ErrPermanent = errors.New("permanent")

// Handler processes one message body. Returning nil acks it.
type Handler func(ctx context.Context, body []byte) error

const retryDelay = 3 * time.Second

// Consume runs until ctx is cancelled, reconnecting whenever the connection
// drops, so the owning service keeps serving HTTP while RabbitMQ is down.
// Transient handler errors are requeued after retryDelay; permanent ones are
// dead-lettered.
func Consume(ctx context.Context, url, queue string, handle Handler) {
	for ctx.Err() == nil {
		if err := consumeOnce(ctx, url, queue, handle); err != nil {
			log.Printf("broker: consumer %s: %v (retrying in %s)", queue, err, retryDelay)
		}
		select {
		case <-ctx.Done():
		case <-time.After(retryDelay):
		}
	}
}

func consumeOnce(ctx context.Context, url, queue string, handle Handler) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := declareTopology(ch); err != nil {
		return err
	}
	// Handlers run one at a time, so hold at most one unacked message.
	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.ConsumeWithContext(ctx, queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	log.Printf("broker: consuming %s", queue)
	for d := range deliveries {
		err := handle(ctx, d.Body)
		switch {
		case err == nil:
			_ = d.Ack(false)
		case errors.Is(err, ErrPermanent):
			log.Printf("broker: %s message %s dead-lettered: %v", queue, d.MessageId, err)
			_ = d.Nack(false, false)
		default:
			log.Printf("broker: %s message %s failed, will retry: %v", queue, d.MessageId, err)
			select {
			case <-ctx.Done():
			case <-time.After(retryDelay):
			}
			_ = d.Nack(false, true)
		}
	}
	if ctx.Err() != nil {
		return nil
	}
	return errors.New("delivery channel closed")
}
