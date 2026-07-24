package graphql

import (
	"context"
	"testing"
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
