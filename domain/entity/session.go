package entity

import "fmt"

// Session はセッションを表します。
type Session struct {
	ID          string
	Name        string
	CreatorID   string
	QueueHead   int
	StateType   string
	QueueTracks []*QueueTrack
}

// MoveToPlay はセッションのStateTypeをPlayに状態遷移します。ｓ
func (s *Session) MoveToPlay() error {
	if s.StateType == Pause.String() || s.StateType == Stop.String() {
		s.StateType = Play.String()
		return nil
	}
	return fmt.Errorf("state type from %s to Play: %w", s.StateType, ErrChangeSessionStateNotPermit)
}
