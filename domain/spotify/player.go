//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package spotify

import (
	"context"

	"github.com/camphor-/relaym-server/domain/entity"
)

// Player はSpotifyの曲の操作に関連するAPIを呼び出すためのインターフェースです。
type Player interface {
	CurrentlyPlaying(ctx context.Context) (*entity.CurrentPlayingInfo, error)
	Play(ctx context.Context, deviceID string) error
	PlayWithTracks(ctx context.Context, deviceID string, trackURIs []string) error
	Pause(ctx context.Context, deviceID string) error
	AddToQueue(ctx context.Context, trackURI string, deviceID string) error
	SetRepeatMode(ctx context.Context, on bool, deviceID string) error
	SetShuffleMode(ctx context.Context, on bool, deviceID string) error
	SkipAllTracks(ctx context.Context, deviceID string, trackURI string) error
}
