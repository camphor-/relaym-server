//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package repository

import "github.com/camphor-/relaym-server/domain/entity"

// QueueTrack はsessionのqueueを操作するためのリポジトリです。
type QueueTrack interface {
}
