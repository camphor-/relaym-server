package entity

// QueueTrack はsessionに属するqueue内の曲を表します。
type QueueTrack struct {
	Index     int
	URI       string
	SessionID string
}
