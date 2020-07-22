package entity

import "github.com/google/uuid"

// QueueTrackToStore はsessionに属するqueue内に曲を挿入する際に使用します
type QueueTrackToStore struct {
	ID        string
	URI       string
	SessionID string
}

// QueueTrack はsessionに属するqueue内の曲を表します。
type QueueTrack struct {
	Index     int
	URI       string
	SessionID string
}

// NewQueueTrackToStore はQueueTrackToStateのポインタを生成する関数です。
func NewQueueTrackToStore(uri string, sessionID string) *QueueTrackToStore {
	return &QueueTrackToStore{
		ID:        uuid.New().String(),
		URI:       uri,
		SessionID: sessionID,
	}
}
