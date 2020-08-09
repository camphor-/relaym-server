package usecase

import (
	"context"
	"fmt"

	"github.com/camphor-/relaym-server/log"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/camphor-/relaym-server/domain/spotify"
)

// SessionUseCase はセッションに関するユースケースです。
type SessionUseCase struct {
	sessionRepo repository.Session
	userRepo    repository.User
	playerCli   spotify.Player
	trackCli    spotify.TrackClient
	userCli     spotify.User
	pusher      event.Pusher
	timerUC     *SessionTimerUseCase
}

// NewSessionUseCase はSessionUseCaseのポインタを生成します。
func NewSessionUseCase(sessionRepo repository.Session, userRepo repository.User, playerCli spotify.Player, trackCli spotify.TrackClient, userCli spotify.User, pusher event.Pusher, timerUC *SessionTimerUseCase) *SessionUseCase {
	return &SessionUseCase{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		playerCli:   playerCli,
		trackCli:    trackCli,
		userCli:     userCli,
		pusher:      pusher,
		timerUC:     timerUC,
	}
}

// EnqueueTrack はセッションのqueueにTrackを追加します。
func (s *SessionUseCase) EnqueueTrack(ctx context.Context, sessionID string, trackURI string) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("FindByID sessionID=%s: %w", sessionID, err)
	}

	err = s.sessionRepo.StoreQueueTrack(ctx, &entity.QueueTrackToStore{
		URI:       trackURI,
		SessionID: sessionID,
	})
	if err != nil {
		return fmt.Errorf("StoreQueueTrack URI=%s, sessionID=%s: %w", trackURI, sessionID, err)
	}

	if session.ShouldCallEnqueueAPINow() {
		err = s.playerCli.Enqueue(ctx, trackURI, session.DeviceID)
		if err != nil {
			return fmt.Errorf("Enqueue URI=%s, sessionID=%s: %w", trackURI, sessionID, err)
		}
	}
	s.pusher.Push(&event.PushMessage{
		SessionID: sessionID,
		Msg:       entity.EventAddTrack,
	})

	return nil
}

// CreateSession は与えられたセッション名のセッションを作成します。
func (s *SessionUseCase) CreateSession(ctx context.Context, sessionName string, creatorID string, allowToControlByOthers bool) (*entity.SessionWithUser, error) {
	creator, err := s.userRepo.FindByID(creatorID)
	if err != nil {
		return nil, fmt.Errorf("FindByID userID=%s: %w", creatorID, err)
	}

	newSession, err := entity.NewSession(sessionName, creatorID, allowToControlByOthers)
	if err != nil {
		return nil, fmt.Errorf("NewSession sessionName=%s: %w", sessionName, err)
	}

	err = s.sessionRepo.StoreSession(ctx, newSession)
	if err != nil {
		return nil, fmt.Errorf("StoreSession sessionName=%s: %w", sessionName, err)
	}
	return entity.NewSessionWithUser(newSession, creator), nil
}

// CanConnectToPusher はイベントをクライアントにプッシュするためのコネクションを貼れるかどうかチェックします。
func (s *SessionUseCase) CanConnectToPusher(ctx context.Context, sessionID string) error {
	sess, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	// セッションが再生中なのに同期チェックがされていなかったら始める
	// サーバ再起動でタイマーがなくなると、イベントが正しくクライアントに送られなくなるのでこのタイミングで復旧させる。
	if exists := s.timerUC.existsTimer(sessionID); !exists && sess.IsPlaying() {
		fmt.Printf("session timer not found: create timer: sessionID=%s\n", sessionID)
		go s.timerUC.startTrackEndTrigger(ctx, sessionID)
	}

	return nil
}

// SetDevice は指定されたidのセッションの作成者と再生する端末を紐付けて再生するデバイスを指定します。
func (s *SessionUseCase) SetDevice(ctx context.Context, sessionID string, deviceID string) error {
	sess, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	sess.DeviceID = deviceID
	if err := s.sessionRepo.Update(ctx, sess); err != nil {
		return fmt.Errorf("update device id: device_id=%s session_id=%s: %w", deviceID, sess.ID, err)
	}

	return nil
}

// GetSession は指定されたidからsessionの情報を返します
func (s *SessionUseCase) GetSession(ctx context.Context, sessionID string) (*entity.SessionWithUser, []*entity.Track, *entity.CurrentPlayingInfo, error) {
	logger := log.New()
	session, err := s.sessionRepo.FindByIDForUpdate(ctx, sessionID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("FindByID sessionID=%s: %w", sessionID, err)
	}

	creator, err := s.userRepo.FindByID(session.CreatorID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("FindByID userID=%s: %w", session.CreatorID, err)
	}

	trackURIs := make([]string, len(session.QueueTracks))
	for i, queueTrack := range session.QueueTracks {
		trackURIs[i] = queueTrack.URI
	}

	tracks, err := s.trackCli.GetTracksFromURI(ctx, trackURIs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get tracks: track_uris=%s: %w", trackURIs, err)
	}

	cpi, err := s.playerCli.CurrentlyPlaying(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("CurrentlyPlaying: %w", err)
	}

	if !s.timerUC.existsTimer(sessionID) {
		return entity.NewSessionWithUser(session, creator), tracks, cpi, nil
	}

	isExpired, err := s.timerUC.isTimerExpired(sessionID)

	if isExpired {
		return entity.NewSessionWithUser(session, creator), tracks, cpi, nil
	}

	if err != nil {
		return nil, nil, nil, fmt.Errorf("isTimerExpired: %w", err)
	}

	if err := session.IsPlayingCorrectTrack(cpi); err != nil {
		logger.Infoj(map[string]interface{}{
			"message": "IsPlayingCorrectTrack failed from getSession",
		})
		s.timerUC.deleteTimer(session.ID)
		s.timerUC.handleInterrupt(session)

		if updateErr := s.sessionRepo.Update(ctx, session); updateErr != nil {
			return nil, nil, nil, fmt.Errorf("update session id=%s: %v: %w", session.ID, err, updateErr)
		}
	}

	return entity.NewSessionWithUser(session, creator), tracks, cpi, nil
}

// GetActiveDevices はログインしているユーザがSpotifyを起動している端末を取得します。
func (s *SessionUseCase) GetActiveDevices(ctx context.Context) ([]*entity.Device, error) {
	return s.userCli.GetActiveDevices(ctx)
}
