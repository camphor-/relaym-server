//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package spotify

import (
	"context"

	"github.com/camphor-/relaym-server/domain/entity"
)

// User はSpotifyのユーザに関連したAPIを呼び出すためのインターフェイスです。
type User interface {
	GetMe(ctx context.Context) (*entity.SpotifyUser, error)
	GetActiveDevices(ctx context.Context) (*entity.Devices, error)
}
