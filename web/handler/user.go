package handler

import (
	"net/http"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/log"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/labstack/echo/v4"
)

// UserHandler は /users 以下のエンドポイントを管理する構造体です。
type UserHandler struct {
	userUC *usecase.UserUseCase
}

// NewUserHandler はUserHandlerのポインタを生成する関数です。
func NewUserHandler(userUC *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

// GetMe は GET /users/me に対応するハンドラーです。
func (h *UserHandler) GetMe(c echo.Context) error {
	logger := log.New()

	ctx := c.Request().Context()

	user, su, err := h.userUC.GetMe(ctx)
	if err != nil {
		logger.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, &userRes{
		ID:          user.ID,
		URI:         user.SpotifyURI(),
		DisplayName: user.DisplayName,
		IsPremium:   su.Premium(),
	})
}

type userRes struct {
	ID          string `json:"id"`
	URI         string `json:"url"`
	DisplayName string `json:"display_name"`
	IsPremium   bool   `json:"is_premium"`
}

func toDeviceJSON(devices []*entity.Device) []*deviceJSON {
	deviceJSONs := make([]*deviceJSON, len(devices))

	for i, device := range devices {
		deviceJSONs[i] = &deviceJSON{
			ID:           device.ID,
			IsRestricted: device.IsRestricted,
			Name:         device.Name,
		}
	}
	return deviceJSONs
}

type devicesRes struct {
	Devices []*deviceJSON `json:"devices"`
}

type deviceJSON struct {
	ID           string `json:"id"`
	IsRestricted bool   `json:"is_restricted"`
	Name         string `json:"name"`
}
