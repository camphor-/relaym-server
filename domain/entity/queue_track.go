package entity

// QueueTrack はsessionに属するqueue内の曲を表します。
type QueueTrack struct {
	index     string
	URI       string
	sessionID string
}
