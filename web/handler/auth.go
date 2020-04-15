package handler

import (
	"net/http"

	"github.com/camphor-/relaym-server/usecase"
	"github.com/labstack/echo/v4"
)

// AuthHandler はログインに関連するのエンドポイントを管理する構造体です。
type AuthHandler struct {
	authUC      *usecase.AuthUseCase
	frontendURL string
}

// NewAuthHandler はAuthHandlerのポインタを生成する関数です。
func NewAuthHandler(authUC *usecase.AuthUseCase, frontendURL string) *AuthHandler {
	return &AuthHandler{authUC: authUC, frontendURL: frontendURL}
}

// Login は POST /login に対応するハンドラーです。
func (h *AuthHandler) Login(c echo.Context) error {
	redirectURL := c.QueryParam("redirect_url")

	url := h.authUC.GetAuthURL(redirectURL)

	return c.Redirect(http.StatusFound, url)
}

// Callback はGet /callbackに対応するハンドラーです。
func (h *AuthHandler) Callback(c echo.Context) error {
	if err := c.QueryParam("err"); err != "" {
		c.Logger().Errorf("spotify auth failed: %s", err)
		return c.Redirect(http.StatusFound, h.frontendURL+"?err=spotifyAuthFailed")
	}

	state := c.QueryParam("state")
	if state == "" {
		c.Logger().Errorf("spotify auth failed: state is empty")
		return c.Redirect(http.StatusFound, h.frontendURL+"?err=spotifyAuthFailed")
	}

	code := c.QueryParam("code")
	if code == "" {
		c.Logger().Errorf("spotify auth failed: code is empty")
		return c.Redirect(http.StatusFound, h.frontendURL+"?err=spotifyAuthFailed")
	}

	redirectURL, err := h.authUC.Authorization(state, code)
	if err != nil {
		c.Logger().Errorf("spotify auth failed: %v", err)
		return c.Redirect(http.StatusFound, h.frontendURL+"?err=%spotify auth failed")
	}

	return c.Redirect(http.StatusFound, redirectURL)
}
