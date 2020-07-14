package usecase

import (
	"fmt"

	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/repository"
)

// BatchUseCase はセッションに関するユースケースです。
type BatchUseCase struct {
	sessionRepo repository.Session
	pusher      event.Pusher
}

// NewBatchUseCase はSessionUseCaseのポインタを生成します。
func NewBatchUseCase(sessionRepo repository.Session, pusher event.Pusher) *BatchUseCase {
	return &BatchUseCase{
		sessionRepo: sessionRepo,
		pusher:      pusher,
	}
}

// ArchiveOldSessions は古いSessionのstateをArchivedに変更します
func (s *BatchUseCase) ArchiveOldSessions() error {
	if err := s.sessionRepo.ArchiveSessionsForBatch(); err != nil {
		return fmt.Errorf("call ArchiveSessionsForBatch: %w", err)
	}
	return nil
}
