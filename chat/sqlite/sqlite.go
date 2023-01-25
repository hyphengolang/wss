package repo

import (
	"context"
)

type Chat struct {
	ID []byte
}

type Message struct {
	ID      int
	ChatID  []byte
	Content string
}

type Repo[T any] interface {
	Create(ctx context.Context, t T) error
	FindMany(ctx context.Context) ([]T, error)
	Find(ctx context.Context, key any) (T, error)
	Delete(ctx context.Context, key any) error
}
