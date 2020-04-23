package usecase

import (
	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/spotify"
)

// AuthUseCase は認証・認可に関するユースケースです。
type TrackUseCase struct {
	track spotify.Track
}

// NewAuthUseCase はAuthUseCaseのポインタを生成します。
func NewTrackUseCase(track spotify.Track) *TrackUseCase {
	return &TrackUseCase{track: track}
}

func (t *TrackUseCase) SearckTracks(q string) ([]*entity.Track, error) {
	return t.track.Search(q)
}
