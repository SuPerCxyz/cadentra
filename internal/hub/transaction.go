package hub

import (
	"context"

	"github.com/cadentra/cadentra/internal/store"
)

type transactionalStore interface {
	WithTx(context.Context, func(context.Context) error) error
}

func runMutationTx(ctx context.Context, st store.Store, fn func(context.Context) error) error {
	if txStore, ok := st.(transactionalStore); ok {
		return txStore.WithTx(ctx, fn)
	}
	return fn(ctx)
}
