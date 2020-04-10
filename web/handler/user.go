package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// UserHandler は /users 以下のエンドポイントを管理するハンドラーです
type UserHandler struct {
}

// NewUserHandler はUserHandlerのポインタを生成する関数です。
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// GetMe は GET /users/me に対応するハンドラーです。
func (h *UserHandler) GetMe(c echo.Context) error {
	return c.JSON(http.StatusOK, &userJSON{})
}

type userJSON struct {
	ID          string `json:"id"`
	URI         string `json:"url"`
	DisplayName string `json:"display_name"`
	IsPremium   bool   `json:"is_premium"`
}
