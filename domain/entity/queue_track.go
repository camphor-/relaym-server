package entity

// QueueTrackToStore はsessionに属するqueue内に曲を挿入する際に使用します
type QueueTrackToStore struct {
	URI       string
	SessionID string
}

// QueueTrack はsessionに属するqueue内の曲を表します。
type QueueTrack struct {
	Index     int
	URI       string
	SessionID string
}
