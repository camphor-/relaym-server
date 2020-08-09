package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/camphor-/relaym-server/domain/spotify"
	"github.com/camphor-/relaym-server/log"
)

var waitTimeAfterHandleTrackEnd = 7 * time.Second
var waitTimeAfterHandleSkipTrack = 500 * time.Millisecond

type SessionTimerUseCase struct {
	tm          *entity.SyncCheckTimerManager
	sessionRepo repository.Session
	playerCli   spotify.Player
	pusher      event.Pusher
}

func NewSessionTimerUseCase(sessionRepo repository.Session, playerCli spotify.Player, pusher event.Pusher, tm *entity.SyncCheckTimerManager) *SessionTimerUseCase {
	return &SessionTimerUseCase{tm: tm, sessionRepo: sessionRepo, playerCli: playerCli, pusher: pusher}
}

// startTrackEndTrigger は曲の終了やストップを検知してそれぞれの処理を実行します。 goroutineで実行されることを想定しています。
func (s *SessionTimerUseCase) startTrackEndTrigger(ctx context.Context, sessionID string) {
	// startTrackEndTriggerが終了する際はTimerは必ず不要になるので削除
	defer s.deleteTimer(sessionID)

	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "start track end trigger", "sessionID": sessionID})

	// 曲の再生を待つ
	waitTimer := time.NewTimer(5 * time.Second)
	currentOperation := operationPlay

	triggerAfterTrackEnd := s.tm.CreateExpiredTimer(sessionID)

	for {
		select {
		case <-waitTimer.C:
			logger.Infoj(map[string]interface{}{"message": "waitTimer expired", "sessionID": sessionID})
			if err := s.handleWaitTimerExpired(ctx, sessionID, triggerAfterTrackEnd, currentOperation); err != nil {
				return
			}
		case <-triggerAfterTrackEnd.StopCh():
			logger.Infoj(map[string]interface{}{"message": "stop timer", "sessionID": sessionID})
			waitTimer.Stop()
			return

		case <-triggerAfterTrackEnd.NextCh():
			logger.Debugj(map[string]interface{}{"message": "call to move next track", "sessionID": sessionID})
			waitTimer.Stop()
			triggerAfterTrackEnd.MakeIsTimerExpiredTrue()
			session, err := s.sessionRepo.FindByIDForUpdate(ctx, sessionID)
			if err != nil {
				logger.Errorj(map[string]interface{}{
					"message":   "failed to get session",
					"sessionID": sessionID,
					"error":     err.Error(),
				})
				triggerAfterTrackEnd.UnlockNextCh()
				return
			}
			if session.StateType != entity.Play {
				logger.Errorj(map[string]interface{}{
					"message":   "stateType must be play",
					"sessionID": sessionID,
					"stateType": session.StateType,
				})
				triggerAfterTrackEnd.UnlockNextCh()
				return
			}
			if err := s.playerCli.GoNextTrack(ctx, session.DeviceID); err != nil {
				logger.Errorj(map[string]interface{}{
					"message":   "failed to go next track",
					"sessionID": sessionID,
					"error":     err.Error(),
				})
				triggerAfterTrackEnd.UnlockNextCh()
				return
			}
			nextTrack, err := s.handleTrackEnd(ctx, sessionID)
			if err != nil {
				logger.Errorj(map[string]interface{}{"message": "handleTrackEnd with error", "sessionID": sessionID, "error": err.Error()})
				triggerAfterTrackEnd.UnlockNextCh()
				return
			}
			if !nextTrack {
				logger.Infoj(map[string]interface{}{"message": "no next track", "sessionID": sessionID})
				triggerAfterTrackEnd.UnlockNextCh()
				return
			}

			s.setNewTimerOnWaitTimer(waitTimer, waitTimeAfterHandleSkipTrack)
			logger.Debugj(map[string]interface{}{"message": "reset waitTimer"})
			currentOperation = operationNextTrack

		case <-triggerAfterTrackEnd.ExpireCh():
			triggerAfterTrackEnd.MakeIsTimerExpiredTrue()
			logger.Debugj(map[string]interface{}{"message": "trigger expired", "sessionID": sessionID})
			nextTrack, err := s.handleTrackEnd(ctx, sessionID)
			if err != nil {
				logger.Errorj(map[string]interface{}{"message": "handleTrackEnd with error", "sessionID": sessionID, "error": err.Error()})
				return
			}
			if !nextTrack {
				logger.Infoj(map[string]interface{}{"message": "no next track", "sessionID": sessionID})
				return
			}
			s.setNewTimerOnWaitTimer(waitTimer, waitTimeAfterHandleTrackEnd)
			logger.Debugj(map[string]interface{}{"message": "reset waitTimer"})
			currentOperation = operationNextTrack
		}
	}
}

func (s *SessionTimerUseCase) handleWaitTimerExpired(ctx context.Context, sessionID string, triggerAfterTrackEnd *entity.SyncCheckTimer, currentOperation currentOperation) error {
	_, err := s.sessionRepo.DoInTx(ctx, s.handleWaitTimerExpiredTx(sessionID, triggerAfterTrackEnd, currentOperation))

	return err
}
func (s *SessionTimerUseCase) handleWaitTimerExpiredTx(sessionID string, triggerAfterTrackEnd *entity.SyncCheckTimer, currentOperation currentOperation) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (_ interface{}, returnErr error) {
		logger := log.New()
		logger.Debugj(map[string]interface{}{"message": "currentOperation", "currentOperation": currentOperation})

		playingInfo, err := s.playerCli.CurrentlyPlaying(ctx)
		if err != nil {
			logger.Errorj(map[string]interface{}{
				"message":   "handleWaitTimerExpired: failed to get currently playing info",
				"sessionID": sessionID,
				"error":     err.Error(),
			})
			return nil, fmt.Errorf("failed to get currently playing info")
		}

		sess, err := s.sessionRepo.FindByIDForUpdate(ctx, sessionID)
		if err != nil {
			logger.Errorj(map[string]interface{}{
				"message":   "handleWaitTimerExpired: failed to get session",
				"sessionID": sessionID,
				"error":     err.Error(),
			})
			return nil, fmt.Errorf("failed to get session from repo")
		}

		defer func() {
			if err := s.sessionRepo.Update(ctx, sess); err != nil {
				if returnErr != nil {
					returnErr = fmt.Errorf("update session id=%s: %v: %w", sess.ID, err, returnErr)
				} else {
					returnErr = fmt.Errorf("update session id=%s: %w", sess.ID, err)
				}
			}

			if currentOperation == operationNextTrack {
				triggerAfterTrackEnd.UnlockNextCh()
			}
		}()

		if err := sess.IsPlayingCorrectTrack(playingInfo); err != nil {
			logger.Infoj(map[string]interface{}{
				"message": "IsPlayingCorrectTrack failed from handleWaitTimerExpired",
			})
			s.handleInterrupt(sess)
			return nil, fmt.Errorf("session interrupt")
		}

		track := sess.TrackURIShouldBeAddedWhenHandleTrackEnd()
		if track != "" {
			if err := s.playerCli.Enqueue(ctx, track, sess.DeviceID); err != nil {
				logger.Errorj(map[string]interface{}{
					"message":   "handleWaitTimerExpired: failed to enqueue tracks",
					"sessionID": sessionID,
					"error":     err.Error(),
				})
				s.handleInterrupt(sess)
				return nil, fmt.Errorf("failed to enqueue tracks")
			}
		}

		// ぴったしのタイマーをセットすると、Spotifyでは次の曲の再生が始まってるのにRelaym側では次の曲に進んでおらず、
		// INTERRUPTになってしまう
		remainDuration := playingInfo.Remain() - 2*time.Second

		logger.Infoj(map[string]interface{}{
			"message": "start timer", "sessionID": sessionID, "remainDuration": remainDuration.String(),
		})

		triggerAfterTrackEnd.SetDuration(remainDuration)

		switch currentOperation {
		case operationNextTrack:
			s.pusher.Push(&event.PushMessage{
				SessionID: sess.ID,
				Msg:       entity.NewEventNextTrack(sess.QueueHead),
			})
			//time.Sleep(100 * time.Millisecond)
		}

		return nil, nil
	}
}

// handleTrackEnd はある一曲の再生が終わったときの処理を行います。
func (s *SessionTimerUseCase) handleTrackEnd(ctx context.Context, sessionID string) (bool, error) {

	triggerAfterTrackEndResponse, err := s.sessionRepo.DoInTx(ctx, s.handleTrackEndTx(sessionID))
	if v, ok := triggerAfterTrackEndResponse.(*handleTrackEndResponse); ok {
		// これはトランザクションが失敗してRollbackしたとき
		if err != nil {
			return false, fmt.Errorf("handle track end in transaction: %w", err)
		}
		return v.nextTrack, v.err
	}
	// これはトランザクションが失敗してRollbackしたとき
	if err != nil {
		return false, fmt.Errorf("handle track end in transaction: %w", err)
	}
	return false, nil
}

// handleTrackEndTx はINTERRUPTになってerrorを帰す場合もトランザクションをコミットして欲しいので、
// アプリケーションエラーはhandleTrackEndResponseのフィールドで返すようにしてerrorの返り値はnilにしている
func (s *SessionTimerUseCase) handleTrackEndTx(sessionID string) func(ctx context.Context) (interface{}, error) {
	logger := log.New()
	return func(ctx context.Context) (_ interface{}, returnErr error) {
		sess, err := s.sessionRepo.FindByIDForUpdate(ctx, sessionID)
		if err != nil {
			return &handleTrackEndResponse{nextTrack: false}, fmt.Errorf("find session id=%s: %v", sessionID, err)
		}

		defer func() {
			if err := s.sessionRepo.Update(ctx, sess); err != nil {
				if returnErr != nil {
					returnErr = fmt.Errorf("update session id=%s: %v: %w", sess.ID, err, returnErr)
				} else {
					returnErr = fmt.Errorf("update session id=%s: %w", sess.ID, err)
				}
			}
		}()

		// 曲の再生中にArchivedになった場合
		if sess.StateType == entity.Archived {

			s.pusher.Push(&event.PushMessage{
				SessionID: sess.ID,
				Msg:       entity.EventArchived,
			})

			return &handleTrackEndResponse{
				nextTrack: false,
				err:       nil,
			}, nil
		}

		if err := sess.GoNextTrack(); err != nil && errors.Is(err, entity.ErrSessionAllTracksFinished) {
			s.handleAllTrackFinish(sess)
			return &handleTrackEndResponse{
				nextTrack: false,
				err:       nil,
			}, nil
		}

		logger.Debugj(map[string]interface{}{"message": "next track", "sessionID": sess.ID, "queueHead": sess.QueueHead})

		return &handleTrackEndResponse{nextTrack: true, err: nil}, nil
	}
}

// handleAllTrackFinish はキューの全ての曲の再生が終わったときの処理を行います。
func (s *SessionTimerUseCase) handleAllTrackFinish(sess *entity.Session) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "all track finished", "sessionID": sess.ID})
	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventStop,
	})
}

// handleInterrupt はSpotifyとの同期が取れていないときの処理を行います。
func (s *SessionTimerUseCase) handleInterrupt(sess *entity.Session) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "interrupt detected", "sessionID": sess.ID})

	sess.MoveToStop()

	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventInterrupt,
	})
}

func (s *SessionTimerUseCase) setNewTimerOnWaitTimer(timer *time.Timer, d time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(d)
}

func (s *SessionTimerUseCase) existsTimer(sessionID string) bool {
	_, exists := s.tm.GetTimer(sessionID)
	return exists
}

func (s *SessionTimerUseCase) deleteTimer(sessionID string) {
	s.tm.DeleteTimer(sessionID)
}

func (s *SessionTimerUseCase) isTimerExpired(sessionID string) (bool, error) {
	return s.tm.IsTimerExpired(sessionID)
}

func (s *SessionTimerUseCase) sendToNextCh(sessionID string) {
	go s.tm.SendToNextCh(sessionID)
}

type handleTrackEndResponse struct {
	nextTrack bool
	err       error
}

type currentOperation string

const (
	operationPlay      currentOperation = "play"
	operationNextTrack currentOperation = "NextTrack"
)
