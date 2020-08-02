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
var waitTimeAfterHandleSkipTrack = 300 * time.Millisecond

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
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "start track end trigger", "sessionID": sessionID})

	// 曲の再生を待つ
	waitTimer := time.NewTimer(5 * time.Second)
	currentOperation := operationPlay

	triggerAfterTrackEnd := s.tm.CreateExpiredTimer(sessionID)
	for {
		select {
		case <-waitTimer.C:
			playingInfo, err := s.playerCli.CurrentlyPlaying(ctx)
			if err != nil {
				logger.Errorj(map[string]interface{}{
					"message":   "startTrackEndTrigger: failed to get currently playing info",
					"sessionID": sessionID,
					"error":     err.Error(),
				})
				return
			}

			sess, err := s.sessionRepo.FindByID(ctx, sessionID)
			if err != nil {
				logger.Errorj(map[string]interface{}{
					"message":   "startTrackEndTrigger: failed to get session",
					"sessionID": sessionID,
					"error":     err.Error(),
				})
				return
			}

			if err := sess.IsPlayingCorrectTrack(playingInfo); err != nil {
				s.handleInterrupt(sess)
				if err := s.sessionRepo.Update(ctx, sess); err != nil {
					logger.Errorj(map[string]interface{}{
						"message":   "startTrackEndTrigger: failed to update session after handleInterrupt",
						"sessionID": sessionID,
						"error":     err.Error(),
					})
					return
				}
				return
			}

			logger.Debugj(map[string]interface{}{"message": "currentOperation", "currentOperation": currentOperation})

			switch currentOperation {
			case operationNextTrack:
				s.pusher.Push(&event.PushMessage{
					SessionID: sess.ID,
					Msg:       entity.NewEventNextTrack(sess.QueueHead),
				})
			}

			// ぴったしのタイマーをセットすると、Spotifyでは次の曲の再生が始まってるのにRelaym側では次の曲に進んでおらず、
			// INTERRUPTになってしまう
			remainDuration := playingInfo.Remain() - 2*time.Second

			logger.Infoj(map[string]interface{}{
				"message": "start timer", "sessionID": sessionID, "remainDuration": remainDuration.String(),
			})

			triggerAfterTrackEnd.SetDuration(remainDuration)

		case <-triggerAfterTrackEnd.StopCh():
			logger.Infoj(map[string]interface{}{"message": "stop timer", "sessionID": sessionID})
			waitTimer.Stop()
			s.deleteTimer(sessionID)
			return

		case <-triggerAfterTrackEnd.NextCh():
			logger.Debugj(map[string]interface{}{"message": "call to move next track", "sessionID": sessionID})
			waitTimer.Stop()
			nextTrack, err := s.handleTrackEnd(ctx, sessionID)
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

			waitTimer = time.NewTimer(waitTimeAfterHandleSkipTrack)
			currentOperation = operationNextTrack
			if err != nil {
				logger.Errorj(map[string]interface{}{
					"message":   "startTrackEndTrigger: failed to change string to current operation",
					"sessionID": sessionID,
					"error":     err.Error(),
				})
				return
			}

		case <-triggerAfterTrackEnd.ExpireCh():
			triggerAfterTrackEnd.MakeIsTimerExpiredTrue()
			logger.Debugj(map[string]interface{}{"message": "trigger expired", "sessionID": sessionID})
			nextTrack, err := s.handleTrackEnd(ctx, sessionID)
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
			waitTimer = time.NewTimer(waitTimeAfterHandleTrackEnd)
			currentOperation = operationNextTrack
			if err != nil {
				logger.Errorj(map[string]interface{}{
					"message":   "startTrackEndTrigger: failed to change string to current operation",
					"sessionID": sessionID,
					"error":     err.Error(),
				})
				return
			}
		}
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

		track := sess.TrackURIShouldBeAddedWhenHandleTrackEnd()
		if track != "" {
			if err := s.playerCli.Enqueue(ctx, track, sess.DeviceID); err != nil {
				return &handleTrackEndResponse{
					nextTrack: false,
					err:       fmt.Errorf("call add queue api trackURI=%s: %w", track, err),
				}, nil
			}
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

func (s *SessionTimerUseCase) existsTimer(sessionID string) bool {
	_, exists := s.tm.GetTimer(sessionID)
	return exists
}

func (s *SessionTimerUseCase) stopTimer(sessionID string) {
	s.tm.StopTimer(sessionID)
}

func (s *SessionTimerUseCase) deleteTimer(sessionID string) {
	s.tm.DeleteTimer(sessionID)
}

func (s *SessionTimerUseCase) isTimerExpired(sessionID string) (bool, error) {
	return s.tm.IsTimerExpired(sessionID)
}

func (s *SessionTimerUseCase) sendToNextCh(sessionID string) error {
	return s.tm.SendToNextCh(sessionID)
}

type handleTrackEndResponse struct {
	nextTrack bool
	err       error
}

type CurrentOperation string

const (
	operationPlay      CurrentOperation = "play"
	operationNextTrack CurrentOperation = "NextTrack"
)
