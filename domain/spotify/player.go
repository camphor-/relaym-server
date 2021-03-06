//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package spotify

import (
	"context"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
)

// Player はSpotifyの曲の操作に関連するAPIを呼び出すためのインターフェースです。
type Player interface {
	CurrentlyPlaying(ctx context.Context) (*entity.CurrentPlayingInfo, error)
	PlayWithTracksAndPosition(ctx context.Context, deviceID string, trackURIs []string, position time.Duration) error
	Pause(ctx context.Context, deviceID string) error
	Enqueue(ctx context.Context, trackURI string, deviceID string) error
	SetRepeatMode(ctx context.Context, on bool, deviceID string) error
	SetShuffleMode(ctx context.Context, on bool, deviceID string) error
	DeleteAllTracksInQueue(ctx context.Context, deviceID string, trackURI string) error
	GoNextTrack(ctx context.Context, deviceID string) error
}
