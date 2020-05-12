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

var errorNumDuplicateEntry uint16 = 1062

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
		Name:      dto.Name,
		CreatorID: dto.CreatorID,
		QueueHead: dto.QueueHead,
		StateType: dto.StateType,
	}, nil
}

func (r *SessionRepository) StoreSession(session *entity.Session) error {
	dto := &sessionDTO{
		ID:        session.ID,
		Name:      session.Name,
		CreatorID: session.CreatorID,
		QueueHead: session.QueueHead,
		StateType: session.StateType,
	}

	if err := r.dbMap.Insert(dto); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == errorNumDuplicateEntry {
			return fmt.Errorf("insert session: %w", entity.ErrSessionAlreadyExisted)
		}
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *SessionRepository) StoreQueueTrack(queueTrack *entity.QueueTrackToStore) error {
	if _, err := r.dbMap.Exec("INSERT INTO queue_tracks SELECT MAX(qt.index)+1, ?, ? from queue_tracks as qt;", queueTrack.URI, queueTrack.SessionID); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == errorNumDuplicateEntry {
			return fmt.Errorf("insert queue_tracks: %w", entity.ErrQueueTrackAlreadyExisted)
		}
		return fmt.Errorf("insert queue_tracks: %w", err)
	}
	return nil
}

type sessionDTO struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	CreatorID string `db:"creator_id"`
	QueueHead int    `db:"queue_head"`
	StateType string `db:"state_type"`
}

type queueTrackDTO struct {
	Index     int    `db:"index"`
	URI       string `db:"uri"`
	SessionID string `db:"session_id"`
}
