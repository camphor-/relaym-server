package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/camphor-/relaym-server/domain/service"
)

// UserUseCase はユーザに関係するアプリケーションロジックを担当する構造体です。
type UserUseCase struct {
	userRepo repository.User
}

// NewUserUseCase はUserUseCaseのポインタを生成する関数です。
func NewUserUseCase(userRepo repository.User) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

// GetMe はログインしているユーザを取得します。
func (u *UserUseCase) GetMe(ctx context.Context) (*entity.User, error) {
	id, ok := service.GetUserIDFromContext(ctx)
	if !ok {
		return nil, errors.New("get user id from context")
	}
	user, err := u.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("find user from repo id=%s: %v", id, user)
	}
	return user, nil
}
