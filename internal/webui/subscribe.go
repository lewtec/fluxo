package webui

import (
	"context"

	"github.com/lucasew/fluxo/internal/session"
)

// subOutBuffer absorbs brief SSE client stalls. When full, sendSub drops the
// oldest value so this goroutine keeps draining the EventBus channel — a blocked
// send would fill the bus buffer and drop unrelated events (e.g. torrent_removed).
const subOutBuffer = 16

// sendSub delivers v without blocking on a slow consumer. Returns false when
// ctx is already cancelled and the caller should exit. If the buffer is full,
// the oldest value is dropped so this goroutine can keep draining EventBus.
func sendSub[T any](ctx context.Context, out chan T, v T) bool {
	if ctx.Err() != nil {
		return false
	}
	select {
	case out <- v:
		return true
	default:
	}
	// Drop oldest to make room (same freshness preference as EventBus.Publish).
	select {
	case <-out:
	default:
	}
	if ctx.Err() != nil {
		return false
	}
	select {
	case out <- v:
	default:
		// Still full (racing with another send or consumer); skip this value.
	}
	return true
}

// subscribeFilter wires EventBus → buffered out channel. match returns
// (value, true) to emit an event and false to skip. The out channel is closed
// when ctx is cancelled or the bus subscription ends.
func subscribeFilter[T any](
	ctx context.Context,
	bus *session.EventBus,
	match func(session.Event) (T, bool),
) <-chan T {
	subID, eventChan := bus.Subscribe()
	out := make(chan T, subOutBuffer)

	go func() {
		defer bus.Unsubscribe(subID)
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-eventChan:
				if !ok {
					return
				}
				v, ok := match(event)
				if !ok {
					continue
				}
				if !sendSub(ctx, out, v) {
					return
				}
			}
		}
	}()

	return out
}
