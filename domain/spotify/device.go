//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package spotify

import (
	"context"

	"github.com/camphor-/relaym-server/domain/entity"
)

// Device はSpotifyのデバイスに関連したAPIを呼び出すためのインターフェイスです。
type DeviceClient interface {
	GetActiveDevices(ctx context.Context) (*entity.Devices, error)
}
