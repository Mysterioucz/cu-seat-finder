package registry

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc/resolver"
)

// A gRPC resolver for consul://<name> targets, so grpc.NewClient("consul://library")
// keeps an up-to-date list of healthy Library instances and the round_robin
// balancer spreads calls over them. Registered on import.
func init() { resolver.Register(builder{}) }

type builder struct{}

func (builder) Scheme() string { return Scheme }

func (builder) Build(t resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	name := t.URL.Host // consul://library
	if name == "" {
		name = t.Endpoint() // consul:///library
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &grpcResolver{name: name, cc: cc, cancel: cancel, now: make(chan struct{}, 1)}
	go r.watch(ctx)
	return r, nil
}

type grpcResolver struct {
	name   string
	cc     resolver.ClientConn
	cancel context.CancelFunc
	now    chan struct{}
	once   sync.Once
}

// watch re-reads the healthy instances every 2s (or when gRPC asks after a
// connection failure) and hands them to the balancer.
func (r *grpcResolver) watch(ctx context.Context) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		lctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		addrs, err := fetch(lctx, r.name)
		cancel()
		if err != nil {
			r.cc.ReportError(err)
		} else {
			state := resolver.State{}
			for _, a := range addrs {
				state.Addresses = append(state.Addresses, resolver.Address{Addr: a})
				state.Endpoints = append(state.Endpoints, resolver.Endpoint{Addresses: []resolver.Address{{Addr: a}}})
			}
			if len(addrs) == 0 {
				r.cc.ReportError(ErrNoInstance)
			} else {
				r.cc.UpdateState(state)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		case <-r.now:
		}
	}
}

func (r *grpcResolver) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case r.now <- struct{}{}:
	default:
	}
}

func (r *grpcResolver) Close() { r.once.Do(r.cancel) }
