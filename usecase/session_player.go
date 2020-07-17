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
)

// SessionStateUseCase はセッションの再生に関するユースケースです。
type SessionStateUseCase struct {
	sessionRepo repository.Session
	playerCli   spotify.Player
	pusher      event.Pusher
	timerUC     *SessionTimerUseCase
}

// NewSessionPlayerUseCase はSessionPlayerUseCaseのポインタを生成します。
func NewSessionStateUseCase(sessionRepo repository.Session, playerCli spotify.Player, pusher event.Pusher, timerUC *SessionTimerUseCase) *SessionStateUseCase {
	return &SessionStateUseCase{sessionRepo: sessionRepo, playerCli: playerCli, pusher: pusher, timerUC: timerUC}
}

// ChangeSessionState は与えられたセッションのstateを操作します。
func (s *SessionStateUseCase) ChangeSessionState(ctx context.Context, sessionID string, st entity.StateType) error {
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
func (s *SessionStateUseCase) playORResume(ctx context.Context, sess *entity.Session) error {
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

	go s.timerUC.startTrackEndTrigger(ctx, sess.ID)

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

func (s *SessionStateUseCase) stopToPlay(ctx context.Context, sess *entity.Session) error {
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
		if err := s.playerCli.Enqueue(ctx, trackURIs[i], sess.DeviceID); err != nil {
			return fmt.Errorf("call add queue api trackURI=%s: %w", trackURIs[i], err)
		}
	}
	return nil
}

// Pause はセッションのstateをPLAY→PAUSEに変更して曲の再生を一時停止します。
func (s *SessionStateUseCase) pause(ctx context.Context, sess *entity.Session) error {
	if err := s.playerCli.Pause(ctx, sess.DeviceID); err != nil && !errors.Is(err, entity.ErrActiveDeviceNotFound) {
		return fmt.Errorf("call pause api: %w", err)
	}

	s.timerUC.stopTimer(sess.ID)

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
func (s *SessionStateUseCase) archive(ctx context.Context, session *entity.Session) error {
	switch session.StateType {
	case entity.Play:
		if err := s.playerCli.Pause(ctx, session.DeviceID); err != nil && !errors.Is(err, entity.ErrActiveDeviceNotFound) {
			return fmt.Errorf("call pause api: %w", err)
		}
	case entity.Archived:
		return nil
	}

	s.timerUC.deleteTimer(session.ID)

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
func (s *SessionStateUseCase) stop(session *entity.Session) error {
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

func (s *SessionStateUseCase) archiveToStop(session *entity.Session) error {
	session.MoveToStop()

	threeDaysAfter := time.Now().AddDate(0, 0, 3).UTC()
	if err := s.sessionRepo.UpdateWithExpiredAt(session, threeDaysAfter); err != nil {
		return fmt.Errorf("update session id=%s: %w", session.ID, err)
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: session.ID,
		Msg:       entity.EventUnarchive,
	})

	return nil
}
