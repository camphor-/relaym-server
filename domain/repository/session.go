//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package repository

import "github.com/camphor-/relaym-server/domain/entity"

// Session はsessionを管理するためのリポジトリです。
type Session interface {
	FindByID(id string) (*entity.Session, error)
	Store(*entity.Session) error
}
