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
func (u *UserUseCase) GetMe(ctx context.Context) (*entity.User, *entity.SpotifyUser, error) {
	id, ok := service.GetUserIDFromContext(ctx)
	if !ok {
		return nil, nil, errors.New("get user id from context")
	}
	user, err := u.userRepo.FindByID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("find user from repo id=%s: %w", id, err)
	}

	su, err := u.userCli.GetMe(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("get me through spotify client id=%s : %w", id, err)
	}

	return user, su, nil
}
