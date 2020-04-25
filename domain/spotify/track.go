package spotify

import (
	"context"

	"github.com/camphor-/relaym-server/domain/entity"
)

type Track interface {
	Search(ctx context.Context, q string) ([]*entity.Track, error)
}
