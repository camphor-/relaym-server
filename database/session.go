package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"

	"github.com/go-gorp/gorp/v3"
	"github.com/go-sql-driver/mysql"
)

var _ repository.Session = &SessionRepository{}

// SessionRepository は repository.SessionRepository を満たす構造体です
type SessionRepository struct {
	dbMap *gorp.DbMap
}

// NewSessionRepository はSessionRepositoryのポインタを生成する関数です
func NewSessionRepository(dbMap *gorp.DbMap) *SessionRepository {
	dbMap.AddTableWithName(sessionDTO{}, "sessions")
	dbMap.AddTableWithName(queueTrackDTO{}, "queue_tracks")
	return &SessionRepository{dbMap: dbMap}
}

// FindByID は指定されたIDを持つsessionをDBから取得します
func (r *SessionRepository) FindByID(id string) (*entity.Session, error) {
	var dto sessionDTO
	if err := r.dbMap.SelectOne(&dto, "SELECT id, name, creator_id, queue_head, state_type FROM sessions WHERE id = ?", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("select session: %w", entity.ErrSessionNotFound)
		}
		return nil, fmt.Errorf("select session: %w", err)
	}
	return &entity.Session{
		ID:        dto.ID,
		Name:      dto.name,
		CreatorID: dto.creator_id,
		QueueHead: dto.queue_head,
		StateType: dto.state_type,
	}, nil
}

func (r *SessionRepository) StoreSessions(session *entity.Session) error {
	dto := &sessionDTO{
		ID:         session.ID,
		name:       session.Name,
		creator_id: session.CreatorID,
		queue_head: session.QueueHead,
		state_type: session.StateType,
	}

	if err := r.dbMap.Insert(dto); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("insert session: %w", entity.ErrSessionAlreadyExisted)
		}
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *SessionRepository) StoreQueueTracks(queueTrack *entity.QueueTrack) error {

	if tx, err := r.dbMap.Begin(); err != nil {
		return fmt.Errorf("gorp.DbMap.Begin() error: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if max_idx, err := tx.SelectInt("SELECT MAX(index) AS index FROM queue_tracks"); err != nil {
		tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("select session: %w", entity.ErrSessionNotFound)
		}
		return fmt.Errorf("select session: %w", err)
	}

	dto := &queueTrackDTO{
		index:      max_idx + 1,
		uri:        queueTrack.URI,
		session_id: queueTrack.SessionID,
	}

	if err := tx.dbMap.Insert(dto); err != nil {
		tx.Rollback()
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("insert queue: %w", entity.ErrQueueAlreadyExisted)
		}
		return fmt.Errorf("insert queue: %w", err)
	}

	return tx.Commit().Error
}

type sessionDTO struct {
	ID         string `db:"id"`
	name       string `db:"name"`
	creator_id string `db:"creator_id"`
	queue_head int    `db:"queue_head"`
	state_type string `db:"state_type"`
}

type queueTrackDTO struct {
	index      int    `db:"index"`
	uri        string `db:"uri"`
	session_id string `db:"session_id"`
}
