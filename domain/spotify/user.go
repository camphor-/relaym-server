package spotify

import (
	"context"

	"github.com/camphor-/relaym-server/domain/entity"
)

// User はSpotifyのユーザに関連したAPIを呼び出すためのインターフェイスです。
type User interface {
	GetMe(ctx context.Context) (*entity.SpotifyUser, error)
}