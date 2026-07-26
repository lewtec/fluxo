package graphql

import (
	"context"
	"testing"
	"time"

	"github.com/lucasew/fluxo/internal/session"
)

func TestSendSubDeliversWhenRoom(t *testing.T) {
	ctx := t.Context()
	out := make(chan int, 2)

	if !sendSub(ctx, out, 1) {
		t.Fatal("expected send success")
	}
	if !sendSub(ctx, out, 2) {
		t.Fatal("expected send success")
	}

	if got := <-out; got != 1 {
		t.Fatalf("got %d want 1", got)
	}
	if got := <-out; got != 2 {
		t.Fatalf("got %d want 2", got)
	}
}

func TestSendSubDropsOldestWhenFull(t *testing.T) {
	ctx := t.Context()
	out := make(chan int, 2)

	if !sendSub(ctx, out, 1) || !sendSub(ctx, out, 2) {
		t.Fatal("fill failed")
	}
	// Buffer full: next send should drop 1 and keep 2,3.
	if !sendSub(ctx, out, 3) {
		t.Fatal("expected send success with drop-oldest")
	}

	var got []int
drain:
	for {
		select {
		case v := <-out:
			got = append(got, v)
		default:
			break drain
		}
	}
	if len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("got %v want [2 3]", got)
	}
}

func TestSendSubReturnsFalseWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	out := make(chan int, 1)
	if sendSub(ctx, out, 1) {
		t.Fatal("expected false when ctx already cancelled")
	}
	select {
	case <-out:
		t.Fatal("expected no value delivered after cancel")
	default:
	}
}

func TestSubscribeFilterMatchesAndSkips(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	bus := session.NewEventBus()
	defer bus.Close()

	out := subscribeFilter(ctx, bus, func(event session.Event) (string, bool) {
		if event.Type != session.EventTorrentRemoved {
			return "", false
		}
		return event.ID, true
	})

	bus.Publish(session.Event{Type: session.EventTorrentAdded, ID: "skip-me"})
	bus.Publish(session.Event{Type: session.EventTorrentRemoved, ID: "t-removed"})

	select {
	case got := <-out:
		if got != "t-removed" {
			t.Fatalf("got %q want t-removed", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for matched event")
	}

	// Unmatched events must not leak a value.
	select {
	case got := <-out:
		t.Fatalf("unexpected extra value %q", got)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestSubscribeFilterStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	bus := session.NewEventBus()
	defer bus.Close()

	out := subscribeFilter(ctx, bus, func(event session.Event) (string, bool) {
		return event.ID, true
	})

	cancel()

	select {
	case _, ok := <-out:
		if ok {
			// May or may not have a value; channel must close soon.
			_, ok2 := <-out
			if ok2 {
				t.Fatal("expected out closed after cancel")
			}
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for out to close after cancel")
	}
}
