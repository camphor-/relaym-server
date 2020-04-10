package database

import "github.com/camphor-/relaym-server/domain/entity"

// UserRepository は repository.UserRepository を満たす構造体です
type UserRepository struct {
}

// NewUserRepository はUserRepositoryのポインタを生成する関数です
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// FindByID は指定されたIDを持つユーザをDBから取得します
func (r *UserRepository) FindByID(id string) (*entity.User, error) {
	return entity.NewUser(id), nil // TODO : 実際にDBから取得する
}
