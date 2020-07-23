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

var syncCheckOffset = 5 * time.Second

type SessionTimerUseCase struct {
	tm          *entity.SyncCheckTimerManager
	sessionRepo repository.Session
	playerCli   spotify.Player
	pusher      event.Pusher
}

func NewSessionTimerUseCase(sessionRepo repository.Session, playerCli spotify.Player, pusher event.Pusher) *SessionTimerUseCase {
	return &SessionTimerUseCase{tm: entity.NewSyncCheckTimerManager(), sessionRepo: sessionRepo, playerCli: playerCli, pusher: pusher}
}

// startTrackEndTrigger は曲の終了やストップを検知してそれぞれの処理を実行します。 goroutineで実行されることを想定しています。
func (s *SessionTimerUseCase) startTrackEndTrigger(ctx context.Context, sessionID string) {
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
func (s *SessionTimerUseCase) handleTrackEnd(ctx context.Context, sessionID string) (triggerAfterTrackEnd *entity.SyncCheckTimer, nextTrack bool, returnErr error) {
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
		if err := s.playerCli.Enqueue(ctx, track, sess.DeviceID); err != nil {
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
func (s *SessionTimerUseCase) handleAllTrackFinish(sess *entity.Session) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "all track finished", "sessionID": sess.ID})
	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventStop,
	})
}

// handleInterrupt はSpotifyとの同期が取れていないときの処理を行います。
func (s *SessionTimerUseCase) handleInterrupt(sess *entity.Session) error {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "interrupt detected", "sessionID": sess.ID})

	s.tm.StopTimer(sess.ID)

	sess.MoveToStop()

	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventInterrupt,
	})
	return nil
}

func (s *SessionTimerUseCase) existsTimer(sessionID string) bool {
	_, exists := s.tm.GetTimer(sessionID)
	return exists
}

func (s *SessionTimerUseCase) stopTimer(sessionID string) {
	s.tm.StopTimer(sessionID)
}

func (s *SessionTimerUseCase) deleteTimer(sessionID string) {
	s.tm.StopTimer(sessionID)
}