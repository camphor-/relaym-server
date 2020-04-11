package usecase

import (
	"fmt"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"
)

// UserUseCase はユーザに関係するアプリケーションロジックを担当する構造体です。
type UserUseCase struct {
	userRepo repository.User
}

// NewUserUseCase はUseUseCaseのポインタを生成する関数です。
func NewUserUseCase(userRepo repository.User) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

// GetByID は渡されたユーザIDを持つユーザを取得します。
func (u *UserUseCase) GetByID(id string) (*entity.User, error) {
	user, err := u.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("find user from repo id=%s: %v", id, user)
	}
	return user, nil
}
