//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package spotify

import "context"

// Player はSpotifyの曲の操作に関連するAPIを呼び出すためのインターフェースです。
type Player interface {
	CurrentlyPlayiwng(ctx context.Context) (bool, error)
	Play(ctx context.Context) error
	Pause(ctx context.Context) error
	AddToQueue(ctx context.Context, trackID string) error
	SetRepeatMode(ctx context.Context, on bool) error
}
