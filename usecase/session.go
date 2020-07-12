package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/camphor-/relaym-server/domain/spotify"
	"github.com/camphor-/relaym-server/log"
)

var syncCheckOffset = 5 * time.Second

// SessionUseCase はセッションに関するユースケースです。
type SessionUseCase struct {
	tm          *entity.SyncCheckTimerManager
	sessionRepo repository.Session
	userRepo    repository.User
	playerCli   spotify.Player
	trackCli    spotify.TrackClient
	userCli     spotify.User
	pusher      event.Pusher
}

// NewSessionUseCase はSessionUseCaseのポインタを生成します。
func NewSessionUseCase(sessionRepo repository.Session, userRepo repository.User, playerCli spotify.Player, trackCli spotify.TrackClient, userCli spotify.User, pusher event.Pusher) *SessionUseCase {
	return &SessionUseCase{
		tm:          entity.NewSyncCheckTimerManager(),
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		playerCli:   playerCli,
		trackCli:    trackCli,
		userCli:     userCli,
		pusher:      pusher,
	}
}

// AddQueueTrack はセッションのqueueにTrackを追加します。
func (s *SessionUseCase) AddQueueTrack(ctx context.Context, sessionID string, trackURI string) error {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("FindByID sessionID=%s: %w", sessionID, err)
	}

	err = s.sessionRepo.StoreQueueTrack(&entity.QueueTrackToStore{
		URI:       trackURI,
		SessionID: sessionID,
	})
	if err != nil {
		return fmt.Errorf("StoreQueueTrack URI=%s, sessionID=%s: %w", trackURI, sessionID, err)
	}

	if session.ShouldCallAddQueueAPINow() {
		err = s.playerCli.AddToQueue(ctx, trackURI, session.DeviceID)
		if err != nil {
			return fmt.Errorf("AddToQueue URI=%s, sessionID=%s: %w", trackURI, sessionID, err)
		}
	}
	s.pusher.Push(&event.PushMessage{
		SessionID: sessionID,
		Msg:       entity.EventAddTrack,
	})

	return nil
}

// CreateSession は与えられたセッション名のセッションを作成します。
func (s *SessionUseCase) CreateSession(sessionName string, creatorID string) (*entity.SessionWithUser, error) {
	creator, err := s.userRepo.FindByID(creatorID)
	if err != nil {
		return nil, fmt.Errorf("FindByID userID=%s: %w", creatorID, err)
	}

	newSession, err := entity.NewSession(sessionName, creatorID)
	if err != nil {
		return nil, fmt.Errorf("NewSession sessionName=%s: %w", sessionName, err)
	}

	err = s.sessionRepo.StoreSession(newSession)
	if err != nil {
		return nil, fmt.Errorf("StoreSession sessionName=%s: %w", sessionName, err)
	}
	return entity.NewSessionWithUser(newSession, creator), nil
}

// ChangeSessionState は与えられたセッションのstateを操作します。
func (s *SessionUseCase) ChangeSessionState(ctx context.Context, sessionID string, st entity.StateType) error {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	if !session.IsValidNextStateFromAPI(st) {
		return fmt.Errorf("state type from %s to %s: %w", session.StateType, st, entity.ErrChangeSessionStateNotPermit)
	}

	switch st {
	case entity.Play:
		if err := s.playORResume(ctx, session); err != nil {
			return fmt.Errorf("playORResume sessionID=%s: %w", sessionID, err)
		}
	case entity.Pause:
		if err := s.pause(ctx, session); err != nil {
			return fmt.Errorf("pause sessionID=%s: %w", sessionID, err)
		}
	case entity.Archived:
		if err := s.archive(ctx, session); err != nil {
			return fmt.Errorf("archive sessionID=%s: %w", sessionID, err)
		}
	case entity.Stop:
		if err := s.stop(session); err != nil {
			return fmt.Errorf("unarchive sessionID=%s: %w", sessionID, err)
		}
	}
	return nil
}

// playORResume はセッションのstateを STOP, PAUSE → PLAY に変更して曲の再生を始めます。
func (s *SessionUseCase) playORResume(ctx context.Context, sess *entity.Session) error {
	// active device not foundになった場合、スマホ側でSpotifyアプリを強制的に開かせてactiveにするので、正常処理をする。
	// その処理を切り替えるフラグとして使う。
	var returnErr error

	userID, _ := service.GetUserIDFromContext(ctx)

	err := s.playerCli.SetRepeatMode(ctx, false, sess.DeviceID)
	if errors.Is(err, entity.ErrActiveDeviceNotFound) && sess.IsCreator(userID) {
		returnErr = err
	} else if err != nil {
		return fmt.Errorf("call set repeat off api: %w", err)
	}

	err = s.playerCli.SetShuffleMode(ctx, false, sess.DeviceID)
	if errors.Is(err, entity.ErrActiveDeviceNotFound) && sess.IsCreator(userID) {
		returnErr = err
	} else if err != nil {
		return fmt.Errorf("call set repeat off api: %w", err)
	}

	if sess.IsResume(entity.Play) {
		err := s.playerCli.Play(ctx, sess.DeviceID)
		if errors.Is(err, entity.ErrActiveDeviceNotFound) && sess.IsCreator(userID) {
			returnErr = err
		} else if err != nil {
			return fmt.Errorf("call play api: %w", err)
		}
	} else {
		if err := s.stopToPlay(ctx, sess); err != nil {
			return fmt.Errorf("start to play: %w", err)
		}
	}

	if err := sess.MoveToPlay(); err != nil {
		return fmt.Errorf("move to play id=%s: %w", sess.ID, err)
	}

	if err := s.sessionRepo.Update(sess); err != nil {
		return fmt.Errorf("update session id=%s: %w", sess.ID, err)
	}

	go s.startTrackEndTrigger(ctx, sess.ID)

	// nilじゃない場合にイベントを送ってしまうと、client側が GET /sessions/:id を叩いてしまい、
	// 即座に INTERRUPT が発火されてしまって困るので条件分岐する。
	if returnErr == nil {
		s.pusher.Push(&event.PushMessage{
			SessionID: sess.ID,
			Msg:       entity.EventPlay,
		})
	}

	return returnErr
}

func (s *SessionUseCase) stopToPlay(ctx context.Context, sess *entity.Session) error {
	trackURIs, err := sess.TrackURIsShouldBeAddedWhenStopToPlay()
	if err != nil {
		return fmt.Errorf(": %w", err)
	}

	if err := s.playerCli.SkipAllTracks(ctx, sess.DeviceID, trackURIs[0]); err != nil {
		return fmt.Errorf("call SkipAllTracks: %w", err)
	}
	for i := 0; i < len(trackURIs); i++ {
		if i == 0 {
			if err := s.playerCli.PlayWithTracks(ctx, sess.DeviceID, trackURIs[:1]); err != nil {
				return fmt.Errorf("call play api with tracks %v: %w", trackURIs[:1], err)
			}
			continue
		}
		if err := s.playerCli.AddToQueue(ctx, trackURIs[i], sess.DeviceID); err != nil {
			return fmt.Errorf("call add queue api trackURI=%s: %w", trackURIs[i], err)
		}
	}
	return nil
}

// Pause はセッションのstateをPLAY→PAUSEに変更して曲の再生を一時停止します。
func (s *SessionUseCase) pause(ctx context.Context, sess *entity.Session) error {
	if err := s.playerCli.Pause(ctx, sess.DeviceID); err != nil && !errors.Is(err, entity.ErrActiveDeviceNotFound) {
		return fmt.Errorf("call pause api: %w", err)
	}

	s.tm.StopTimer(sess.ID)

	if err := sess.MoveToPause(); err != nil {
		return fmt.Errorf("move to pause id=%s: %w", sess.ID, err)
	}

	if err := s.sessionRepo.Update(sess); err != nil {
		return fmt.Errorf("update session id=%s: %w", sess.ID, err)
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventPause,
	})

	return nil
}

// archive はセッションのstateをARCHIVEDに変更します。
func (s *SessionUseCase) archive(ctx context.Context, session *entity.Session) error {
	switch session.StateType {
	case entity.Play:
		if err := s.playerCli.Pause(ctx, session.DeviceID); err != nil && !errors.Is(err, entity.ErrActiveDeviceNotFound) {
			return fmt.Errorf("call pause api: %w", err)
		}
	case entity.Archived:
		return nil
	}

	s.tm.DeleteTimer(session.ID)

	session.MoveToArchived()

	if err := s.sessionRepo.Update(session); err != nil {
		return fmt.Errorf("update session id=%s: %w", session.ID, err)
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: session.ID,
		Msg:       entity.EventArchived,
	})

	return nil
}

// stop はセッションのstateをSTOPに変更します。
func (s *SessionUseCase) stop(session *entity.Session) error {
	switch session.StateType {
	case entity.Stop:
		return nil
	case entity.Archived:
		if err := s.archiveToStop(session); err != nil {
			return fmt.Errorf("call archiveToCall: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("state type from %s to STOP: %w", session.StateType, entity.ErrChangeSessionStateNotPermit)
	}
}

func (s *SessionUseCase) archiveToStop(session *entity.Session) error {
	session.MoveToStop()

	currentDateTime := time.Now()

	if err := s.sessionRepo.UpdateWithExpiredAt(session, &currentDateTime); err != nil {
		return fmt.Errorf("update session id=%s: %w", session.ID, err)
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: session.ID,
		Msg:       entity.EventUnarchive,
	})

	return nil
}

// CanConnectToPusher はイベントをクライアントにプッシュするためのコネクションを貼れるかどうかチェックします。
func (s *SessionUseCase) CanConnectToPusher(ctx context.Context, sessionID string) error {
	sess, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	// セッションが再生中なのに同期チェックがされていなかったら始める
	// サーバ再起動でタイマーがなくなると、イベントが正しくクライアントに送られなくなるのでこのタイミングで復旧させる。
	if _, ok := s.tm.GetTimer(sessionID); !ok && sess.IsPlaying() {
		fmt.Printf("session timer not found: create timer: sessionID=%s\n", sessionID)
		go s.startTrackEndTrigger(ctx, sessionID)
	}

	return nil
}

// startTrackEndTrigger は曲の終了やストップを検知してそれぞれの処理を実行します。 goroutineで実行されることを想定しています。
func (s *SessionUseCase) startTrackEndTrigger(ctx context.Context, sessionID string) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "start track end trigger", "sessionID": sessionID})

	time.Sleep(7 * time.Second) // 曲の再生が始まるのを待つ
	playingInfo, err := s.playerCli.CurrentlyPlaying(ctx)
	if err != nil {
		logger.Errorj(map[string]interface{}{
			"message":   "startTrackEndTrigger: failed to get currently playing info",
			"sessionID": sessionID,
			"error":     err,
		})
		return
	}
	remainDuration := playingInfo.Remain()

	logger.Infoj(map[string]interface{}{
		"message": "start timer", "sessionID": sessionID, "remainDuration": remainDuration.String(),
	})

	triggerAfterTrackEnd := s.tm.CreateTimer(sessionID, remainDuration+syncCheckOffset)

	for {
		select {
		case <-triggerAfterTrackEnd.StopCh():
			logger.Infoj(map[string]interface{}{"message": "stop timer", "sessionID": sessionID})
			return
		case <-triggerAfterTrackEnd.ExpireCh():
			logger.Debugj(map[string]interface{}{"message": "trigger expired", "sessionID": sessionID})
			timer, nextTrack, err := s.handleTrackEnd(ctx, sessionID)
			if err != nil {
				if errors.Is(err, entity.ErrSessionPlayingDifferentTrack) {
					logger.Infoj(map[string]interface{}{"message": "handleTrackEnd detects interrupt", "sessionID": sessionID, "error": err.Error()})
					return
				}
				logger.Errorj(map[string]interface{}{"message": "handleTrackEnd with error", "sessionID": sessionID, "error": err.Error()})
				return
			}
			if !nextTrack {
				logger.Infoj(map[string]interface{}{"message": "no next track", "sessionID": sessionID})
				return
			}
			triggerAfterTrackEnd = timer
		}
	}
}

// handleTrackEnd はある一曲の再生が終わったときの処理を行います。
func (s *SessionUseCase) handleTrackEnd(ctx context.Context, sessionID string) (triggerAfterTrackEnd *entity.SyncCheckTimer, nextTrack bool, returnErr error) {
	logger := log.New()

	sess, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return nil, false, fmt.Errorf("find session id=%s: %v", sessionID, err)
	}

	defer func() {
		if err := s.sessionRepo.Update(sess); err != nil {
			if returnErr != nil {
				returnErr = fmt.Errorf("update session id=%s: %v: %w", sess.ID, err, returnErr)
			} else {
				returnErr = fmt.Errorf("update session id=%s: %w", sess.ID, err)
			}
		}
	}()

	if sess.StateType == entity.Archived {
		s.tm.DeleteTimer(sessionID)

		s.pusher.Push(&event.PushMessage{
			SessionID: sessionID,
			Msg:       entity.EventArchived,
		})

		return nil, false, nil
	}

	if err := sess.GoNextTrack(); err != nil && errors.Is(err, entity.ErrSessionAllTracksFinished) {
		s.handleAllTrackFinish(sess)
		return nil, false, nil
	}

	track := sess.TrackURIShouldBeAddedWhenHandleTrackEnd()
	if track != "" {
		if err := s.playerCli.AddToQueue(ctx, track, sess.DeviceID); err != nil {
			return nil, false, fmt.Errorf("call add queue api trackURI=%s: %w", track, err)
		}
	}

	logger.Debugj(map[string]interface{}{"message": "next track", "sessionID": sessionID, "queueHead": sess.QueueHead})

	playingInfo, err := s.playerCli.CurrentlyPlaying(ctx)
	if err != nil {
		if errors.Is(err, entity.ErrActiveDeviceNotFound) {
			s.tm.DeleteTimer(sess.ID)
			if interErr := s.handleInterrupt(sess); interErr != nil {
				returnErr = fmt.Errorf("handle interrupt: %w", interErr)
				return nil, false, returnErr
			}
			returnErr = err
			return nil, false, returnErr
		}
		returnErr = fmt.Errorf("get currently playing info id=%s: %v", sessionID, err)
		return nil, false, returnErr
	}

	if err := sess.IsPlayingCorrectTrack(playingInfo); err != nil {
		s.tm.DeleteTimer(sess.ID)
		if interErr := s.handleInterrupt(sess); interErr != nil {
			returnErr = fmt.Errorf("check whether playing correct track: handle interrupt: %v: %w", interErr, err)
			return nil, false, returnErr
		}
		returnErr = fmt.Errorf("check whether playing correct track: %w", err)
		return nil, false, returnErr
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: sessionID,
		Msg:       entity.NewEventNextTrack(sess.QueueHead),
	})
	triggerAfterTrackEnd = s.tm.CreateTimer(sessionID, playingInfo.Remain()+syncCheckOffset)

	logger.Infoj(map[string]interface{}{
		"message": "restart timer", "sessionID": sessionID, "remainDuration": playingInfo.Remain().String(),
	})

	return triggerAfterTrackEnd, true, nil

}

// handleAllTrackFinish はキューの全ての曲の再生が終わったときの処理を行います。
func (s *SessionUseCase) handleAllTrackFinish(sess *entity.Session) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "all track finished", "sessionID": sess.ID})
	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventStop,
	})
}

// handleInterrupt はSpotifyとの同期が取れていないときの処理を行います。
func (s *SessionUseCase) handleInterrupt(sess *entity.Session) error {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "interrupt detected", "sessionID": sess.ID})

	sess.MoveToStop()

	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventInterrupt,
	})
	return nil
}

// SetDevice は指定されたidのセッションの作成者と再生する端末を紐付けて再生するデバイスを指定します。
func (s *SessionUseCase) SetDevice(ctx context.Context, sessionID string, deviceID string) error {
	sess, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	sess.DeviceID = deviceID
	if err := s.sessionRepo.Update(sess); err != nil {
		return fmt.Errorf("update device id: device_id=%s session_id=%s: %w", deviceID, sess.ID, err)
	}

	return nil
}

// GetSession は指定されたidからsessionの情報を返します
func (s *SessionUseCase) GetSession(ctx context.Context, sessionID string) (*entity.SessionWithUser, []*entity.Track, *entity.CurrentPlayingInfo, error) {
	session, err := s.sessionRepo.FindByID(sessionID)
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

	if err := session.IsPlayingCorrectTrack(cpi); err != nil {
		s.tm.StopTimer(sessionID)
		if interErr := s.handleInterrupt(session); interErr != nil {
			return nil, nil, nil, fmt.Errorf("check whether playing correct track: handle interrupt: %v: %w", interErr, err)
		}

		if updateErr := s.sessionRepo.Update(session); updateErr != nil {
			return nil, nil, nil, fmt.Errorf("update session id=%s: %v: %w", session.ID, err, updateErr)
		}

		return entity.NewSessionWithUser(session, creator), tracks, cpi, nil
	}
	return entity.NewSessionWithUser(session, creator), tracks, cpi, nil
}

// GetActiveDevices はログインしているユーザがSpotifyを起動している端末を取得します。
func (s *SessionUseCase) GetActiveDevices(ctx context.Context) ([]*entity.Device, error) {
	return s.userCli.GetActiveDevices(ctx)
}
