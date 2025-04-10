package events

import (
	"app/logging"
	"context"
	"go.uber.org/zap"
)

var log = logging.Get().Named("dispatcher")

// eventHandler defines a function type that processes an event of a generic type E within a given context and returns an error.
type eventHandler[E any] func(ctx context.Context, event E) error

// listeners holds a collection of event handlers to be invoked during event dispatching.
var listeners []any

// Handle adds the specified event handler to the list of listeners for processing events.
func Handle[E any](handler eventHandler[E]) {
	listeners = append(listeners, handler)
}

// Dispatch sends the given event to all registered listeners that match the event type. Returns an error if a handler fails.
func Dispatch[E any](ctx context.Context, event E) error {
	log.Debug("dispatching event", zap.Any("event", event))

	for _, listener := range listeners {
		if h, ok := listener.(eventHandler[E]); ok {
			err := h(ctx, event)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
