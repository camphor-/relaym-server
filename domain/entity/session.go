package entity

import (
	"fmt"

	"github.com/camphor-/relaym-server/log"

	"github.com/google/uuid"
)

// Session はセッションを表します。
type Session struct {
	ID          string
	Name        string
	CreatorID   string
	DeviceID    string
	StateType   StateType
	QueueHead   int
	QueueTracks []*QueueTrack
}

type SessionWithUser struct {
	*Session
	Creator *User
}

// NewSession はSessionのポインタを生成する関数です。
func NewSession(name string, creatorID string) (*Session, error) {
	return &Session{
		ID:          uuid.New().String(),
		Name:        name,
		CreatorID:   creatorID,
		DeviceID:    "",
		StateType:   Stop,
		QueueHead:   0,
		QueueTracks: nil,
	}, nil
}

// NewSessionWithUser はSession(ポインタ)からSessionWithUser(ポインタ)を生成します
func NewSessionWithUser(session *Session, creator *User) *SessionWithUser {
	return &SessionWithUser{
		Session: &Session{
			ID:          session.ID,
			Name:        session.Name,
			CreatorID:   creator.ID,
			DeviceID:    "",
			StateType:   session.StateType,
			QueueHead:   session.QueueHead,
			QueueTracks: session.QueueTracks,
		},
		Creator: creator,
	}
}

// MoveToPlay はセッションのStateTypeをPlayに状態遷移します。
func (s *Session) MoveToPlay() error {
	if err := s.canMoveFromStopToPlay(); s.StateType == Stop && err != nil {
		return fmt.Errorf("can not to move to play: %w", err)
	}

	s.StateType = Play
	return nil
}

// MoveToPause はセッションのStateTypeをPauseに状態遷移します。
func (s *Session) MoveToPause() error {
	if s.StateType == Play || s.StateType == Pause {
		s.StateType = Pause
		return nil
	}
	return fmt.Errorf("state type from %s to Pause: %w", s.StateType, ErrChangeSessionStateNotPermit)
}

// MoveToStop はセッションのStateTypeをStopに状態遷移します。
func (s *Session) MoveToStop() error {
	if s.StateType == Pause {
		return fmt.Errorf("state type from Pause to Stop: %w", ErrChangeSessionStateNotPermit)
	}
	s.StateType = Stop
	return nil
}

// IsCreator は指定されたユーザがセッションの作成者かどうか返します。
func (s *Session) IsCreator(userID string) bool {
	return s.CreatorID == userID
}

// GoNextTrack 次の曲の状態に進めます。
func (s *Session) GoNextTrack() error {
	if len(s.QueueTracks) <= s.QueueHead+1 {
		s.QueueHead++ // https://github.com/camphor-/relaym-server/blob/master/docs/definition.md#%E7%8F%BE%E5%9C%A8%E5%AF%BE%E8%B1%A1%E3%81%AE%E6%9B%B2%E3%81%AE%E3%82%A4%E3%83%B3%E3%83%87%E3%83%83%E3%82%AF%E3%82%B9-head
		s.StateType = Stop
		return ErrSessionAllTracksFinished
	}
	s.QueueHead++
	return nil
}

// IsPlayingCorrectTrack は現在の再生状況がセッションの状況と一致しているかチェックします。
func (s *Session) IsPlayingCorrectTrack(playingInfo *CurrentPlayingInfo) error {
	logger := log.New()
	if playingInfo.Track == nil || s.QueueTracks[s.QueueHead].URI != playingInfo.Track.URI {
		logger.Infoj(map[string]interface{}{
			"message":      "session playing different track",
			"queueTrack":   s.QueueTracks[s.QueueHead].URI,
			"playingTrack": playingInfo.Track,
		})
		return fmt.Errorf("session playing different track: queue track %s, but playing track %v: %w", s.QueueTracks[s.QueueHead].URI, playingInfo.Track, ErrSessionPlayingDifferentTrack)
	}
	return nil
}

// ShouldCallAddQueueAPINow は今すぐキューに追加するAPIを叩くかどうか判定します。
// 最初の再生開始時(Stop→Play時)は一気にキューに追加するけど、それ以外のときは随時追加したいので、
// それをチェックするために使います。
func (s *Session) ShouldCallAddQueueAPINow() bool {
	return s.StateType == Play || s.StateType == Pause
}

// IsResume は次のStateTypeへの移行がポーズからの再開かどうかを返します。
func (s *Session) IsResume(nextState StateType) bool {
	return s.StateType == Pause && nextState == Play
}

// TrackURIsShouldBeAddedWhenStartPlay は再生を開始するときにSpotifyのキューに追加するTrackURIを抽出します。
func (s *Session) TrackURIsShouldBeAddedWhenStopToPlay() ([]string, error) {
	if err := s.canMoveFromStopToPlay(); err != nil {
		return []string{}, fmt.Errorf("can not to move to play: %w", err)
	}

	uris := make([]string, len(s.QueueTracks)-s.QueueHead)
	for i := 0; i < len(s.QueueTracks)-s.QueueHead; i++ {
		trackIndex := i + s.QueueHead
		uris[i] = s.QueueTracks[trackIndex].URI
	}
	return uris, nil
}

// canMoveFromStopToPlay はセッションのStateTypeをPlayに状態遷移しても良いかどうか返します。
func (s *Session) canMoveFromStopToPlay() error {
	if s.StateType != Stop {
		return fmt.Errorf("state type from %s to Pause: %w", s.StateType, ErrChangeSessionStateNotPermit)
	}
	if s.isEmptyQueue() {
		return ErrQueueTrackNotFound
	}

	if len(s.QueueTracks) == s.QueueHead {
		return ErrNextQueueTrackNotFound
	}

	return nil
}

// IsPlaying は現在のStateTypeがPlayかどうか返します。
func (s *Session) IsPlaying() bool {
	return s.StateType == Play
}

// isEmptyQueue はキューが空かどうか返します。
func (s *Session) isEmptyQueue() bool {
	return len(s.QueueTracks) == 0
}

type StateType string

const (
	Play  StateType = "PLAY"
	Pause StateType = "PAUSE"
	Stop  StateType = "STOP"
)

var stateTypes = []StateType{Play, Pause, Stop}

// NewStateType はstringから対応するStateTypeを生成します。
func NewStateType(stateType string) (StateType, error) {
	for _, st := range stateTypes {
		if st.String() == stateType {
			return st, nil
		}
	}
	return "", fmt.Errorf("stateType = %s:%w", stateType, ErrInvalidStateType)
}

// String はfmt.Stringerを満たすメソッドです。
func (st StateType) String() string {
	return string(st)
}
