package broker

import (
	"errors"
	"fmt"
	"testing"
)

func TestRequeue(t *testing.T) {
	transient := errors.New("db down")
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"transient is retried", transient, true},
		{"permanent is dead-lettered", Permanent(transient), false},
		{"wrapped permanent is dead-lettered", fmt.Errorf("handler: %w", Permanent(transient)), false},
	}
	for _, c := range cases {
		if got := requeue(c.err); got != c.want {
			t.Errorf("%s: requeue = %v, want %v", c.name, got, c.want)
		}
	}
	if !errors.Is(Permanent(transient), transient) {
		t.Error("Permanent should unwrap to the original error")
	}
}
