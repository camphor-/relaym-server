//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package repository

import (
	"github.com/camphor-/relaym-server/domain/entity"
	"golang.org/x/oauth2"
)

// Session はsessionを管理するためのリポジトリです。
type Session interface {
	FindByID(id string) (*entity.Session, error)
	StoreSession(*entity.Session) error
	Update(*entity.Session) error
	UpdateWithTimeStamp(*entity.Session) error
	StoreQueueTrack(*entity.QueueTrackToStore) error
	FindCreatorTokenBySessionID(string) (*oauth2.Token, string, error)
	ArchiveSessionsForBatch() error
}
