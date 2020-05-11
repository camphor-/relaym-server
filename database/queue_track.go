package database

import (
	"fmt"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"

	"github.com/go-gorp/gorp/v3"
	"github.com/go-sql-driver/mysql"
)

var _ repository.QueueTrack = &QueueTrackRepository{}

// QueueTrackRepository は repository.QueueTrackRepository を満たす構造体です
type QueueTrackRepository struct {
	dbMap *gorp.DbMap
}

// NewQueueTrackRepository はQueueTrackRepositoryのポインタを生成する関数です
func NewQueueTrackRepository(dbMap *gorp.DbMap) *QueueTrackRepository {
	dbMap.AddTableWithName(queueTrackDTO{}, "queue_tracks")
	return &QueueTrackRepository{dbMap: dbMap}
}

func (r *QueueTrackRepository) Store(queueTrack *entity.QueueTrack) error {
	dto := &queueTrackDTO{
		index:      queueTrack.Index,
		uri:        queueTrack.URI,
		session_id: queueTrack.SessionID,
	}

	if err := r.dbMap.Insert(dto); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("insert queue: %w", entity.ErrQueueAlreadyExisted)
		}
		return fmt.Errorf("insert queue: %w", err)
	}
	return nil
}

type queueTrackDTO struct {
	index      int    `db:"index"`
	uri        string `db:"uri"`
	session_id string `db:"session_id"`
}
