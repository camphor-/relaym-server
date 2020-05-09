package database

import ()

var _ repository.QueueTrack = &QueueTrackRepository{}

// QueueTrackRepository は repository.QueueTrackRepository を満たす構造体です
type QueueTrackRepository struct {
	dbMap *gorp.DbMap
}

// NewQueueTrackRepository はQueueTrackRepositoryのポインタを生成する関数です
func NewQueueTrackRepository(dbMap *gorp.DbMap) *QueueTrackRepository {
	dbMap.AddTableWithName(queueTrackDTO{}, "queue_tracks")
	return &QueueTrackRepository{dbMap: dbMap}
}
