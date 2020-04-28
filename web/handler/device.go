package handler

import (
	"net/http"

	"github.com/camphor-/relaym-server/usecase"

	"github.com/labstack/echo/v4"
)

// DeviceHandler は /users 以下のエンドポイントを管理する構造体です。
type DeviceHandler struct {
	deviceUC *usecase.DeviceUsecase
}

// NewUserHandler はUserHandlerのポインタを生成する関数です。
func NewDeviceHandler(deviceUC *usecase.DeviceUseCase) *UserHandler {
	return &DeviceHandler{deviceUC: deviceUC}
}

// GetActiveDevices は GET /users/me/devices に対応するハンドラーです。
func (h *DeviceHandler) GetActiveDevices(c echo.Context) error {
	return nil
}
