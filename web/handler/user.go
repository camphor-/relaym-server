package handler

import (
	"net/http"

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
	ctx := c.Request().Context()

	user, err := h.userUC.GetMe(ctx)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, &userRes{
		ID:          user.ID,
		URI:         user.SpotifyURI(),
		DisplayName: user.DisplayName,
		IsPremium:   false, // TODO : 正しくPremium情報を取得する
	})
}

func (h *UserHandler) GetActiveDevices(c echo.Context) error {
	return nil
}

type userRes struct {
	ID          string `json:"id"`
	URI         string `json:"url"`
	DisplayName string `json:"display_name"`
	IsPremium   bool   `json:"is_premium"`
}
