package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/camphor-/relaym-server/domain/spotify"
)

// UserUseCase はユーザに関係するアプリケーションロジックを担当する構造体です。
type UserUseCase struct {
	userRepo repository.User
	userCli  spotify.User
}

// NewUserUseCase はUserUseCaseのポインタを生成する関数です。
func NewUserUseCase(userCli spotify.User, userRepo repository.User) *UserUseCase {
	return &UserUseCase{userRepo: userRepo, userCli: userCli}
}

// GetMe はログインしているユーザを取得します。
func (u *UserUseCase) GetMe(ctx context.Context) (*entity.User, error) {
	id, ok := service.GetUserIDFromContext(ctx)
	if !ok {
		return nil, errors.New("get user id from context")
	}
	user, err := u.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("find user from repo id=%s: %w", id, err)
	}
	return user, nil
}

// GetActiveDevices はログインしているユーザがSpotifyを起動している端末を取得します。
func (u *UserUseCase) GetActiveDevices(ctx context.Context) ([]*entity.Device, error) {
	return u.userCli.GetActiveDevices(ctx)
}
