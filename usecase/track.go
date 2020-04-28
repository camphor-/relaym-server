package usecase

import (
	"context"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/spotify"
)

// TrackUseCase は音楽に関するユースケースです。
type TrackUseCase struct {
	trackCli spotify.TrackClient
}

// NewTrackUseCase はTrackUseCaseのポインタを生成します。
func NewTrackUseCase(track spotify.TrackClient) *TrackUseCase {
	return &TrackUseCase{trackCli: track}
}

// SearchTracks はクエリから音楽を検索します。
func (t *TrackUseCase) SearckTracks(ctx context.Context, q string) ([]*entity.Track, error) {
	return t.trackCli.Search(ctx, q)
}
