package usecase

import (
	"context"
	"errors"
	"fmt"

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

// NextTrack は指定されたidのsessionを次の曲に進めます
func (s *SessionUseCase) NextTrack(ctx context.Context, sessionID string) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	userID, _ := service.GetUserIDFromContext(ctx)
	if !session.AllowToControlByOthers && !session.IsCreator(userID) {
		return fmt.Errorf("not allowd to control session: %w", entity.ErrSessionNotAllowToControlOthers)
	}

	switch session.StateType {
	case entity.Play:
		if err = s.nextTrackInPlay(ctx, session); err != nil {
			return fmt.Errorf("go next track in play session id=%s: %w", session.ID, err)
		}
	case entity.Pause:
		if err = s.nextTrackInPause(ctx, session); err != nil {
			return fmt.Errorf("go next track in pause session id=%s: %w", session.ID, err)
		}
	case entity.Stop:
		if err = s.nextTrackInStop(ctx, session); err != nil {
			return fmt.Errorf("go next track in stop session id=%s: %w", session.ID, err)
		}
	case entity.Archived:
		return fmt.Errorf("go next track: %w", entity.ErrChangeSessionStateNotPermit)
	}

	return nil
}

// nextTrackInPlay はsessionのstateがPLAYの時のnextTrackの処理を行います
func (s *SessionUseCase) nextTrackInPlay(ctx context.Context, session *entity.Session) error {
	if err := s.playerCli.SkipCurrentTrack(ctx, session.DeviceID); err != nil {
		return fmt.Errorf("SkipCurrentTrack: %w", err)
	}

	// NextChを通してstartTrackEndTriggerに次の曲への繊維を通知
	if err := s.timerUC.sendToNextCh(session.ID); err != nil {
		return fmt.Errorf("ExpiredTimer: %w", err)
	}

	return nil
}

// nextTrackInPause はsessionのstateがPAUSEの時のnextTrackの処理を行います
func (s *SessionUseCase) nextTrackInPause(ctx context.Context, session *entity.Session) error {
	if err := s.playerCli.SkipCurrentTrack(ctx, session.DeviceID); err != nil {
		return fmt.Errorf("SkipCurrentTrack: %w", err)
	}

	if err := session.GoNextTrack(); err != nil && errors.Is(err, entity.ErrSessionAllTracksFinished) {
		s.timerUC.handleAllTrackFinish(session)
		if err := s.sessionRepo.Update(ctx, session); err != nil {
			return fmt.Errorf("update session id=%s: %w", session.ID, err)
		}
		return nil
	}

	// Skipだけだと次の曲の再生が始まってしまう
	if err := s.playerCli.Pause(ctx, session.DeviceID); err != nil {
		return fmt.Errorf("call pause api: %w", err)
	}

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("update session id=%s: %w", session.ID, err)
	}

	track := session.TrackURIShouldBeAddedWhenHandleTrackEnd()
	if track != "" {
		if err := s.playerCli.Enqueue(ctx, track, session.DeviceID); err != nil {
			return fmt.Errorf("enqueue error session id=%s: %w", session.ID, err)
		}
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: session.ID,
		Msg:       entity.NewEventNextTrack(session.QueueHead),
	})

	return nil
}

// nextTrackInStop はsessionのstateがSTOPの時のnextTrackの処理を行います
func (s *SessionUseCase) nextTrackInStop(ctx context.Context, session *entity.Session) error {
	if !session.IsNextTrackExistWhenStateIsStop() {
		return nil
	}

	if err := session.GoNextTrack(); err != nil && errors.Is(err, entity.ErrSessionAllTracksFinished) {
		s.timerUC.handleAllTrackFinish(session)
		if err := s.sessionRepo.Update(ctx, session); err != nil {
			return fmt.Errorf("update session id=%s: %w", session.ID, err)
		}
		return nil
	}

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("update session id=%s: %w", session.ID, err)
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: session.ID,
		Msg:       entity.NewEventNextTrack(session.QueueHead),
	})

	return nil
}

// ChangeSessionState は与えられたセッションのstateを操作します。
func (s *SessionStateUseCase) ChangeSessionState(ctx context.Context, sessionID string, st entity.StateType) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	if !session.IsValidNextStateFromAPI(st) {
		return fmt.Errorf("state type from %s to %s: %w", session.StateType, st, entity.ErrChangeSessionStateNotPermit)
	}

	userID, _ := service.GetUserIDFromContext(ctx)
	if !session.AllowToControlByOthers && !session.IsCreator(userID) {
		return fmt.Errorf("not allowd to control state: %w", entity.ErrSessionNotAllowToControlOthers)
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
		if err := s.stop(ctx, session); err != nil {
			return fmt.Errorf("unarchive sessionID=%s: %w", sessionID, err)
		}
	}
	return nil
}

// playORResume はセッションのstateを STOP, PAUSE → PLAY に変更して曲の再生を始めます。
func (s *SessionStateUseCase) playORResume(ctx context.Context, sess *entity.Session) error {
	if err := s.playerCli.SetRepeatMode(ctx, false, sess.DeviceID); err != nil {
		return fmt.Errorf("call set repeat off api: %w", err)
	}

	if err := s.playerCli.SetShuffleMode(ctx, false, sess.DeviceID); err != nil {
		return fmt.Errorf("call set repeat off api: %w", err)
	}

	if sess.IsResume(entity.Play) {
		if err := s.pauseToPlay(ctx, sess); err != nil {
			return fmt.Errorf("from pause to play: %w", err)
		}
	} else {
		if err := s.stopToPlay(ctx, sess); err != nil {
			return fmt.Errorf("from stop to play: %w", err)
		}
	}

	if err := sess.MoveToPlay(); err != nil {
		return fmt.Errorf("move to play id=%s: %w", sess.ID, err)
	}

	if err := s.sessionRepo.Update(ctx, sess); err != nil {
		return fmt.Errorf("update session id=%s: %w", sess.ID, err)
	}

	go s.timerUC.startTrackEndTrigger(ctx, sess.ID)

	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventPlay,
	})

	return nil
}

func (s *SessionStateUseCase) pauseToPlay(ctx context.Context, sess *entity.Session) error {
	if err := s.playerCli.PlayWithTracksAndPosition(ctx, sess.DeviceID, []string{sess.HeadTrack().URI}, sess.ProgressWhenPaused); err != nil {
		return fmt.Errorf("call play api: %w", err)
	}
	return nil
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
	cpi, err := s.playerCli.CurrentlyPlaying(ctx)
	if err != nil && !errors.Is(err, entity.ErrActiveDeviceNotFound) {
		return fmt.Errorf("call currently playing api: %w", err)
	}
	sess.SetProgressWhenPaused(cpi.Progress)

	if err := s.playerCli.Pause(ctx, sess.DeviceID); err != nil && !errors.Is(err, entity.ErrActiveDeviceNotFound) {
		return fmt.Errorf("call pause api: %w", err)
	}

	s.timerUC.stopTimer(sess.ID)

	if err := sess.MoveToPause(); err != nil {
		return fmt.Errorf("move to pause id=%s: %w", sess.ID, err)
	}

	if err := s.sessionRepo.Update(ctx, sess); err != nil {
		return fmt.Errorf("update session id=%s: %w", sess.ID, err)
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventPause,
	})

	return nil
}

// archive はセッションのstateをARCHIVEDに変更します。
// sessionの作成者からのみ呼び出しが可能です
func (s *SessionStateUseCase) archive(ctx context.Context, session *entity.Session) error {
	userID, _ := service.GetUserIDFromContext(ctx)
	if !session.IsCreator(userID) {
		return fmt.Errorf("user is not creator: %w", entity.ErrSessionNotAllowToControlOthers)
	}

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

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("update session id=%s: %w", session.ID, err)
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: session.ID,
		Msg:       entity.EventArchived,
	})

	return nil
}

// stop はセッションのstateをSTOPに変更します。
func (s *SessionStateUseCase) stop(ctx context.Context, session *entity.Session) error {
	userID, _ := service.GetUserIDFromContext(ctx)
	if !session.IsCreator(userID) {
		return fmt.Errorf("user is not creator: %w", entity.ErrSessionNotAllowToControlOthers)
	}

	switch session.StateType {
	case entity.Stop:
		return nil
	case entity.Archived:
		if err := s.archiveToStop(ctx, session); err != nil {
			return fmt.Errorf("call archiveToCall: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("state type from %s to STOP: %w", session.StateType, entity.ErrChangeSessionStateNotPermit)
	}
}

// sessionの作成者からのみ呼び出しが可能です
func (s *SessionStateUseCase) archiveToStop(ctx context.Context, session *entity.Session) error {
	session.MoveToStop()

	session.UpdateExpiredAt()

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("update session id=%s: %w", session.ID, err)
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: session.ID,
		Msg:       entity.EventUnarchive,
	})

	return nil
}
