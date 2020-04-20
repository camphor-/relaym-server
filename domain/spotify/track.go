package spotify

import "github.com/camphor-/relaym-server/domain/entity"

type Track interface {
	Search(q string) ([]*entity.Track, error)
}
