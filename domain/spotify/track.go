//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package spotify

import (
	"context"

	"github.com/camphor-/relaym-server/domain/entity"
)

// TrackClient はSpotifyの音楽に関連したAPIを呼び出すためのインターフェイスです。
type TrackClient interface {
	Search(ctx context.Context, q string) ([]*entity.Track, error)
	GetTracksFromURI(ctx context.Context, trackURIs []string) ([]*entity.Track, error)
}
