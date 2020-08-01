package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"

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
func (r *SessionRepository) FindByID(ctx context.Context, id string) (*entity.Session, error) {
	dao, ok := getTx(ctx)
	if !ok {
		dao = r.dbMap
	}

	var dto sessionDTO
	if err := dao.SelectOne(&dto, "SELECT id, name, creator_id, queue_head, state_type, device_id, expired_at, allow_to_control_by_others, progress_when_paused FROM sessions WHERE id = ?", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("select session: %w", entity.ErrSessionNotFound)
		}
		return nil, fmt.Errorf("select session: %w", err)
	}

	queueTracks, errOnGetQueue := r.getQueueTracksBySessionID(id)
	if errOnGetQueue != nil {
		return nil, fmt.Errorf("get queue tracks: %w", errOnGetQueue)
	}

	stateType, err := entity.NewStateType(dto.StateType)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", entity.ErrInvalidStateType)
	}

	return r.dtoToSession(dto, stateType, queueTracks), nil
}

// FindByIDForUpdate は指定されたIDを持つsessionをDBから取得します
func (r *SessionRepository) FindByIDForUpdate(ctx context.Context, id string) (*entity.Session, error) {
	dao, ok := getTx(ctx)
	if !ok {
		dao = r.dbMap
	}

	var dto sessionDTO
	if err := dao.SelectOne(&dto, "SELECT id, name, creator_id, queue_head, state_type, device_id, expired_at, allow_to_control_by_others, progress_when_paused FROM sessions WHERE id = ? FOR UPDATE", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("select session: %w", entity.ErrSessionNotFound)
		}
		return nil, fmt.Errorf("select session: %w", err)
	}

	queueTracks, errOnGetQueue := r.getQueueTracksBySessionID(id)
	if errOnGetQueue != nil {
		return nil, fmt.Errorf("get queue tracks: %w", errOnGetQueue)
	}

	stateType, err := entity.NewStateType(dto.StateType)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", entity.ErrInvalidStateType)
	}

	return r.dtoToSession(dto, stateType, queueTracks), nil
}

// FindCreatorTokenBySessionID はSessionIDからCreatorのTokenを取得します
func (r *SessionRepository) FindCreatorTokenBySessionID(ctx context.Context, sessionID string) (*oauth2.Token, string, error) {
	dao, ok := getTx(ctx)
	if !ok {
		dao = r.dbMap
	}

	var dto spotifyAuthDTO

	if err := dao.SelectOne(&dto, "SELECT sa.access_token, sa.refresh_token, sa.expiry, sessions.creator_id AS user_id FROM sessions INNER JOIN spotify_auth AS sa ON sa.user_id = sessions.creator_id WHERE sessions.id = ?", sessionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", fmt.Errorf("select session: %w", entity.ErrSessionNotFound)
		}
		return nil, "", fmt.Errorf("select session: %w", err)
	}

	return &oauth2.Token{
		AccessToken:  dto.AccessToken,
		TokenType:    "Bearer",
		RefreshToken: dto.RefreshToken,
		Expiry:       dto.Expiry,
	}, dto.UserID, nil
}

// StoreSession はSessionをDBに挿入します。
func (r *SessionRepository) StoreSession(ctx context.Context, session *entity.Session) error {
	dao, ok := getTx(ctx)
	if !ok {
		dao = r.dbMap
	}

	dto := r.sessionToDTO(session)

	if err := dao.Insert(dto); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == errorNumDuplicateEntry {
			return fmt.Errorf("insert session: %w", entity.ErrSessionAlreadyExisted)
		}
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// Update はセッションの情報を更新します。
func (r *SessionRepository) Update(ctx context.Context, session *entity.Session) error {
	dao, ok := getTx(ctx)
	if !ok {
		dao = r.dbMap
	}

	dto := r.sessionToDTO(session)

	if _, err := dao.Update(dto); err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	return nil
}

// StoreQueueTrack はQueueTrackをDBに挿入します。
func (r *SessionRepository) StoreQueueTrack(ctx context.Context, queueTrack *entity.QueueTrackToStore) error {
	dao, ok := getTx(ctx)
	if !ok {
		dao = r.dbMap
	}

	if _, err := dao.Exec("INSERT INTO queue_tracks(`index`, uri, session_id) SELECT COALESCE(MAX(`index`),-1)+1, ?, ? from queue_tracks as qt WHERE session_id = ?;", queueTrack.URI, queueTrack.SessionID, queueTrack.SessionID); err != nil {
		return fmt.Errorf("insert queue_tracks: %w", err)
	}
	return nil
}

// ArchiveSessionsForBatch は以下の条件に当てはまるSessionのstateをArchivedに変更します
//// - 作成から3日以上が経過している。もしくはArchiveが解除されてから3日以上が経過している
func (r *SessionRepository) ArchiveSessionsForBatch() error {
	currentDateTime := time.Now().UTC()
	if _, err := r.dbMap.Exec("UPDATE sessions SET state_type = 'ARCHIVED' WHERE state_type != 'ARCHIVED' AND expired_at < ?;", currentDateTime); err != nil {
		return fmt.Errorf("update session state_type to ARCHIVED: %w", err)
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

// DoInTx はトランザクションの中でデータベースにアクセスするためのラッパー関数です。
func (r *SessionRepository) DoInTx(ctx context.Context, f func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	tx, err := r.dbMap.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	ctx = context.WithValue(ctx, &txKey, tx)
	v, err := f(ctx)
	if err != nil {
		_ = tx.Rollback()
		return v, fmt.Errorf("rollback: %w", err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return v, fmt.Errorf("failed to commit: rollback: %w", err)
	}
	return v, nil
}

func (r *SessionRepository) dtoToSession(dto sessionDTO, stateType entity.StateType, queueTracks []*entity.QueueTrack) *entity.Session {
	return &entity.Session{
		ID:                     dto.ID,
		Name:                   dto.Name,
		CreatorID:              dto.CreatorID,
		DeviceID:               dto.DeviceID,
		StateType:              stateType,
		QueueHead:              dto.QueueHead,
		QueueTracks:            queueTracks,
		ExpiredAt:              dto.ExpiredAt,
		AllowToControlByOthers: dto.AllowToControlByOthers,
		ProgressWhenPaused:     time.Duration(dto.ProgressWhenPaused) * time.Millisecond,
	}
}

func (r *SessionRepository) sessionToDTO(session *entity.Session) *sessionDTO {
	return &sessionDTO{
		ID:                     session.ID,
		Name:                   session.Name,
		CreatorID:              session.CreatorID,
		QueueHead:              session.QueueHead,
		StateType:              session.StateType.String(),
		DeviceID:               session.DeviceID,
		ExpiredAt:              session.ExpiredAt,
		AllowToControlByOthers: session.AllowToControlByOthers,
		ProgressWhenPaused:     session.ProgressWhenPaused.Milliseconds(),
	}
}

type sessionDTO struct {
	ID                     string    `db:"id"`
	Name                   string    `db:"name"`
	CreatorID              string    `db:"creator_id"`
	QueueHead              int       `db:"queue_head"`
	StateType              string    `db:"state_type"`
	DeviceID               string    `db:"device_id"`
	ExpiredAt              time.Time `db:"expired_at"`
	AllowToControlByOthers bool      `db:"allow_to_control_by_others"`
	ProgressWhenPaused     int64     `db:"progress_when_paused"`
}

type queueTrackDTO struct {
	Index     int    `db:"index"`
	URI       string `db:"uri"`
	SessionID string `db:"session_id"`
}
