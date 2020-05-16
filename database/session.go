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
	dbMap.AddTableWithName(sessionDTO{}, "sessions").SetKeys(false, "ID")
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

	queueTracks, errOnGetQueue := r.getQueueTracksBySessionID(id)
	if errOnGetQueue != nil {
		return nil, fmt.Errorf("get queue tracks: %w", errOnGetQueue)
	}

	return &entity.Session{
		ID:          dto.ID,
		Name:        dto.Name,
		CreatorID:   dto.CreatorID,
		QueueHead:   dto.QueueHead,
		StateType:   dto.StateType,
		QueueTracks: queueTracks,
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

// Update はセッションの情報を更新します。
func (r *SessionRepository) Update(session *entity.Session) error {
	dto := &sessionDTO{
		ID:        session.ID,
		Name:      session.Name,
		CreatorID: session.CreatorID,
		QueueHead: session.QueueHead,
		StateType: session.StateType,
	}

	updateNum, err := r.dbMap.Update(dto)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	if updateNum == 0 {
		return fmt.Errorf("update session: %w", entity.ErrSessionNotFound)
	}
	return nil
}

func (r *SessionRepository) StoreQueueTrack(queueTrack *entity.QueueTrackToStore) error {
	if _, err := r.dbMap.Exec("INSERT INTO queue_tracks(`index`, uri, session_id) SELECT COALESCE(MAX('index'),-1)+1, ?, ? from queue_tracks as qt WHERE session_id = ?;", queueTrack.URI, queueTrack.SessionID, queueTrack.SessionID); err != nil {
		return fmt.Errorf("insert queue_tracks: %w", err)
	}
	return nil
}

func (r *SessionRepository) getQueueTracksBySessionID(id string) ([]*entity.QueueTrack, error) {
	var dto []queueTrackDTO
	if _, err := r.dbMap.Select(&dto, "SELECT * FROM queue_tracks WHERE session_id = ? ORDER BY `index` ASC", id); err != nil {
		return nil, fmt.Errorf("select queue_tracks: %w", err)
	}
	return r.toQueueTracks(dto), nil
}

func (r *SessionRepository) toQueueTracks(resultQueueTracks []queueTrackDTO) []*entity.QueueTrack {
	queueTracks := make([]*entity.QueueTrack, len(resultQueueTracks))

	for i, rs := range resultQueueTracks {
		queueTracks[i] = &entity.QueueTrack{
			Index:     rs.Index,
			URI:       rs.URI,
			SessionID: rs.SessionID,
		}
	}

	return queueTracks
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
