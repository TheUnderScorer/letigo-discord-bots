package events_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"lib/events"
	"testing"
)

type user struct {
	ID string
}

type userCreatedEvent struct {
	user user
}

type userDeletedEvent struct {
	deletedUserID string
}

func TestDispatcher(t *testing.T) {
	t.Run("with one event", func(t *testing.T) {
		testUser := user{
			ID: "1",
		}

		called := false

		events.Handle(func(ctx context.Context, event userCreatedEvent) error {
			called = true
			return nil
		})

		err := events.Dispatch(context.Background(), userCreatedEvent{
			user: testUser,
		})
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("with two event", func(t *testing.T) {
		testUser := user{
			ID: "1",
		}

		calledUserCreatedEvents := 0
		calledUserDeletedEvents := 0

		events.Handle(func(ctx context.Context, event userCreatedEvent) error {
			calledUserCreatedEvents++
			return nil
		})
		events.Handle(func(ctx context.Context, event userDeletedEvent) error {
			calledUserDeletedEvents++
			return nil
		})

		err := events.Dispatch(context.Background(), userCreatedEvent{
			user: testUser,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, calledUserCreatedEvents)
		assert.Equal(t, 0, calledUserDeletedEvents)

		err = events.Dispatch(context.Background(), userDeletedEvent{
			deletedUserID: testUser.ID,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, calledUserCreatedEvents)
		assert.Equal(t, 1, calledUserDeletedEvents)
	})

	t.Run("with error", func(t *testing.T) {
		testUser := user{
			ID: "1",
		}
		errorToReturn := errors.New("test error")

		events.Handle(func(ctx context.Context, event userCreatedEvent) error {
			return errorToReturn
		})

		err := events.Dispatch(context.Background(), userCreatedEvent{
			user: testUser,
		})
		assert.ErrorIs(t, err, errorToReturn)
	})
}
