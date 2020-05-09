package entity

// Session はセッションを表します。
type Session struct {
	ID        string
	Name      string
	CreatorID string
	QueueHead int
	StateType string
}
