//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package repository

import (
	"context"

	"github.com/camphor-/relaym-server/domain/entity"
	"golang.org/x/oauth2"
)

// Session はsessionを管理するためのリポジトリです。
type Session interface {
	FindByID(ctx context.Context, id string) (*entity.Session, error)
	FindByIDForUpdate(ctx context.Context, id string) (*entity.Session, error)
	StoreSession(context.Context, *entity.Session) error
	Update(context.Context, *entity.Session) error
	StoreQueueTrack(context.Context, *entity.QueueTrackToStore) error
	FindCreatorTokenBySessionID(context.Context, string) (*oauth2.Token, string, error)
	ArchiveSessionsForBatch() error
	DoInTx(ctx context.Context, f func(ctx context.Context) (interface{}, error)) (interface{}, error)
}
