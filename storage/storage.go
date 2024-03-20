package storage

import (
	"context"
)

type Storage interface {
	Encode(ctx context.Context, origin string) (index string, err error)
	Decode(ctx context.Context, index string) (origin string, err error)
}
