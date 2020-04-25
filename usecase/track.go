package usecase

import (
	"context"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/spotify"
)

// TrackUseCase は音楽に関するユースケースです。
type TrackUseCase struct {
	track spotify.TrackClient
}

// NewTrackUseCase はTrackUseCaseのポインタを生成します。
func NewTrackUseCase(track spotify.TrackClient) *TrackUseCase {
	return &TrackUseCase{track: track}
}

// SearchTracks はクエリから音楽を検索します。
func (t *TrackUseCase) SearckTracks(ctx context.Context, q string) ([]*entity.Track, error) {
	return t.track.Search(ctx, q)
}
