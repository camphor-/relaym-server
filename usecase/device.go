package usecase

import (
	"context"
	//"errors"
	//"fmt"

	"github.com/camphor-/relaym-server/domain/entity"
	//"github.com/camphor-/relaym-server/domain/service"
)

// DeviceUseCase はデバイスに関係するアプリケーションロジックを担当する構造体です。
type DeviceUseCase struct {
	deviceCli spotify.DeviceClient
}

// NewUserUseCase はUserUseCaseのポインタを生成する関数です。
func NewDeviceUseCase(deviceCli spotify.DeviceClient) *UserUseCase {
	return &DeviceUseCase{deviceCli: deviceCli}
}

// GetMe はログインしているユーザを取得します。
func (u *UserUseCase) GetMe(ctx context.Context) (*entity.User, error) {
	return nil, nil
}
