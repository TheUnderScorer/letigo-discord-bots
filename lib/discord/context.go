package discord

import (
	"context"
	"time"
)

func NewInteractionContext(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	return ctx, cancel
}
