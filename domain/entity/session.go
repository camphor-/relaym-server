package entity

import (
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/log"

	"github.com/google/uuid"
)

// Session はセッションを表します。
type Session struct {
	ID                     string
	Name                   string
	CreatorID              string
	DeviceID               string
	StateType              StateType
	QueueHead              int
	QueueTracks            []*QueueTrack
	ExpiredAt              time.Time
	AllowToControlByOthers bool
	ProgressWhenPaused     time.Duration
}

type SessionWithUser struct {
	*Session
	Creator *User
}

// NewSession はSessionのポインタを生成する関数です。
func NewSession(name string, creatorID string, allowToControlByOthers bool) (*Session, error) {
	return &Session{
		ID:                     uuid.New().String(),
		Name:                   name,
		CreatorID:              creatorID,
		DeviceID:               "",
		StateType:              Stop,
		QueueHead:              0,
		QueueTracks:            nil,
		ExpiredAt:              time.Now().AddDate(0, 0, 3).UTC(),
		AllowToControlByOthers: allowToControlByOthers,
		ProgressWhenPaused:     0 * time.Second,
	}, nil
}

// NewSessionWithUser はSession(ポインタ)からSessionWithUser(ポインタ)を生成します
func NewSessionWithUser(session *Session, creator *User) *SessionWithUser {
	return &SessionWithUser{
		Session: session,
		Creator: creator,
	}
}

// UpdateExpiredAt はexpired_atを現在の時刻から3日後に設定します
func (s *Session) UpdateExpiredAt() {
	threeDaysAfter := time.Now().AddDate(0, 0, 3).UTC()
	s.ExpiredAt = threeDaysAfter
}

// MoveToPlay はセッションのStateTypeをPlayに状態遷移します。
func (s *Session) MoveToPlay() error {
	if err := s.canMoveFromStopToPlay(); s.StateType == Stop && err != nil {
		return fmt.Errorf("can not to move to play: %w", err)
	}

	s.StateType = Play
	s.SetProgressWhenPaused(0 * time.Second)
	return nil
}

// MoveToPause はセッションのStateTypeをPauseに状態遷移します。
func (s *Session) MoveToPause() error {
	if s.StateType == Play || s.StateType == Pause || s.StateType == Stop {
		s.StateType = Pause
		return nil
	}
	return fmt.Errorf("state type from %s to Pause: %w", s.StateType, ErrChangeSessionStateNotPermit)
}

// MoveToStop はセッションのStateTypeをStopに状態遷移します。
func (s *Session) MoveToStop() {
	s.StateType = Stop
	s.SetProgressWhenPaused(0 * time.Second)
}

// MoveToArchived はセッションのStateTypeをArchivedに状態遷移します。
func (s *Session) MoveToArchived() {
	s.StateType = Archived
	s.SetProgressWhenPaused(0 * time.Second)
}

// IsCreator は指定されたユーザがセッションの作成者かどうか返します。
func (s *Session) IsCreator(userID string) bool {
	return s.CreatorID == userID
}

// GoNextTrack 次の曲の状態に進めます。
func (s *Session) GoNextTrack() error {
	s.SetProgressWhenPaused(0 * time.Second)
	if len(s.QueueTracks) <= s.QueueHead+1 {
		s.QueueHead = len(s.QueueTracks) // https://github.com/camphor-/relaym-server/blob/master/docs/definition.md#%E7%8F%BE%E5%9C%A8%E5%AF%BE%E8%B1%A1%E3%81%AE%E6%9B%B2%E3%81%AE%E3%82%A4%E3%83%B3%E3%83%87%E3%83%83%E3%82%AF%E3%82%B9-head
		s.StateType = Stop
		return ErrSessionAllTracksFinished
	}
	s.QueueHead++
	return nil
}

// IsPlayingCorrectTrack は現在の再生状況がセッションの状況と一致しているかチェックします。
func (s *Session) IsPlayingCorrectTrack(playingInfo *CurrentPlayingInfo) error {
	logger := log.New()
	if s.StateType == Stop {
		return nil
	}

	if playingInfo == nil {
		logger.Infoj(map[string]interface{}{
			"message":    "session should be playing but not playing",
			"queueTrack": s.QueueTracks[s.QueueHead].URI,
		})
		return fmt.Errorf("session should be playing: queue track %s, but not playing: %w", s.QueueTracks[s.QueueHead].URI, ErrSessionPlayingDifferentTrack)
	}

	if playingInfo.Track == nil || s.QueueTracks[s.QueueHead].URI != playingInfo.Track.URI {
		logger.Infoj(map[string]interface{}{
			"message":      "session playing different track",
			"queueTrack":   s.QueueTracks[s.QueueHead].URI,
			"playingTrack": playingInfo.Track,
		})
		return fmt.Errorf("session playing different track: queue track %s, but playing track %v: %w", s.QueueTracks[s.QueueHead].URI, playingInfo.Track, ErrSessionPlayingDifferentTrack)
	}

	if playingInfo.Playing != s.IsPlaying() {
		logger.Infoj(map[string]interface{}{
			"message":      "session playing, but spotify is not playing",
			"queueTrack":   s.QueueTracks[s.QueueHead].URI,
			"playingTrack": playingInfo.Track,
		})
		return fmt.Errorf("session playing, but spotify is not playing: %w", ErrSessionPlayingDifferentTrack)

	}

	return nil
}

// ShouldCallEnqueueAPINow は今すぐキューに追加するAPIを叩くかどうか判定します。
// 最後の曲もしくは最後から二番目の曲の再生中に曲を新たに追加された場合はSpotifyのキューに新たに追加したいので、それをチェックするために使います。
func (s *Session) ShouldCallEnqueueAPINow() bool {
	return ((len(s.QueueTracks) - s.QueueHead) < 3) && (s.StateType == Play || s.StateType == Pause)
}

// IsResume は次のStateTypeへの移行がポーズからの再開かどうかを返します。
func (s *Session) IsResume(nextState StateType) bool {
	return s.StateType == Pause && nextState == Play
}

// TrackURIsShouldBeAddedWhenStopToPlay は再生を開始するときにSpotifyのキューに追加するTrackURIを抽出します。
func (s *Session) TrackURIsShouldBeAddedWhenStopToPlay() ([]string, error) {
	if err := s.canMoveFromStopToPlay(); err != nil {
		return []string{}, fmt.Errorf("can not to move to play: %w", err)
	}

	var uris []string
	for i := 0; i < 3; i++ {
		trackIndex := i + s.QueueHead
		uris = append(uris, s.QueueTracks[trackIndex].URI)
		if (len(s.QueueTracks) - s.QueueHead) == i+1 {
			break
		}
	}
	return uris, nil
}

// TrackURIShouldBeAddedWhenHandleTrackEnd はある一曲の再生が終わったときにSpotifyのキューに追加するTrackURIを抽出します。
func (s *Session) TrackURIShouldBeAddedWhenHandleTrackEnd() string {
	if (len(s.QueueTracks) - s.QueueHead) < 3 {
		return ""
	}
	index := s.QueueHead + 2
	return s.QueueTracks[index].URI
}

// canMoveFromStopToPlay はセッションのStateTypeをStopからPlayに状態遷移しても良いかどうか返します。
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

// IsNextTrackExistWhenStateIsStop はstateがstopの時に次の曲が存在するかを調べます
func (s *Session) IsNextTrackExistWhenStateIsStop() bool {
	return len(s.QueueTracks) > s.QueueHead
}

// IsPlaying は現在のStateTypeがPlayかどうか返します。
func (s *Session) IsPlaying() bool {
	return s.StateType == Play
}

// isEmptyQueue はキューが空かどうか返します。
func (s *Session) isEmptyQueue() bool {
	return len(s.QueueTracks) == 0
}

// IsValidNextStateFromAPI は外部からリクエストされたstateの変更の正当性を評価します
func (s *Session) IsValidNextStateFromAPI(nextState StateType) bool {
	if s.StateType == nextState {
		return true
	}

	switch s.StateType {
	case Play:
		return nextState == Pause || nextState == Archived
	case Pause:
		return nextState == Play || nextState == Archived
	case Archived:
		return nextState == Stop
	case Stop:
		return nextState == Play || nextState == Archived
	}

	return false
}

// SetProgressWhenPaused はProgressWhenPausedに時間をセットします。
func (s *Session) SetProgressWhenPaused(d time.Duration) {
	s.ProgressWhenPaused = d
}

// HeadTrack は現在のHeadの曲を返します。
func (s *Session) HeadTrack() *QueueTrack {
	return s.QueueTracks[s.QueueHead]
}

type StateType string

const (
	Play     StateType = "PLAY"
	Pause    StateType = "PAUSE"
	Stop     StateType = "STOP"
	Archived StateType = "ARCHIVED"
)

var stateTypes = []StateType{Play, Pause, Stop, Archived}

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
