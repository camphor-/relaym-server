package database

import (
	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"
	"gopkg.in/gorp.v1"
)

var _ repository.User = &UserRepository{}

// UserRepository は repository.UserRepository を満たす構造体です
type UserRepository struct {
	dbMap *gorp.DbMap
}

// NewUserRepository はUserRepositoryのポインタを生成する関数です
func NewUserRepository(dbMap *gorp.DbMap) *UserRepository {
	return &UserRepository{dbMap: dbMap}
}

// FindByID は指定されたIDを持つユーザをDBから取得します
func (r *UserRepository) FindByID(id string) (*entity.User, error) {
	return entity.NewUser(id), nil // TODO : 実際にDBから取得する
}
