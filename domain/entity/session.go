package entity

import "fmt"

// Session はセッションを表します。
type Session struct {
	ID          string
	Name        string
	CreatorID   string
	QueueHead   int
	StateType   StateType
	QueueTracks []*QueueTrack
}

// MoveToPlay はセッションのStateTypeをPlayに状態遷移します。
func (s *Session) MoveToPlay() error {
	if s.StateType == Pause || s.StateType == Stop {
		s.StateType = Play
		return nil
	}
	return fmt.Errorf("state type from %s to Play: %w", s.StateType, ErrChangeSessionStateNotPermit)
}

// IsCreator は指定されたユーザがセッションの作成者かどうか返します。
func (s *Session) IsCreator(userID string) bool {
	return s.CreatorID == userID
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
