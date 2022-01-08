package handler

import (
	"net/http"

	"github.com/camphor-/relaym-server/config"
	"github.com/camphor-/relaym-server/log"
	"github.com/camphor-/relaym-server/usecase"
	"github.com/labstack/echo/v4"
)

const sevenDays = 60 * 60 * 24 * 7

// AuthHandler はログインに関連するのエンドポイントを管理する構造体です。
type AuthHandler struct {
	authUC      *usecase.AuthUseCase
	frontendURL string
}

// NewAuthHandler はAuthHandlerのポインタを生成する関数です。
func NewAuthHandler(authUC *usecase.AuthUseCase, frontendURL string) *AuthHandler {
	return &AuthHandler{authUC: authUC, frontendURL: frontendURL}
}

// Login は GET /login に対応するハンドラーです。
func (h *AuthHandler) Login(c echo.Context) error {
	logger := log.New()

	redirectURL := c.QueryParam("redirect_url")
	if redirectURL == "" || !config.IsReliableOrigin(redirectURL) {
		redirectURL = h.frontendURL
	}
	url, err := h.authUC.GetAuthURL(redirectURL)
	if err != nil {
		logger.Errorj(map[string]interface{}{"message": "failed to get auth url", "error": err.Error()})
		return c.Redirect(http.StatusFound, redirectURL+"?err=spotifyAuthFailed")
	}

	return c.Redirect(http.StatusFound, url)
}

// Callback はGet /callbackに対応するハンドラーです。
func (h *AuthHandler) Callback(c echo.Context) error {
	logger := log.New()
	if err := c.QueryParam("err"); err != "" {
		logger.Errorj(map[string]interface{}{"message": "spotify auth failed", "error": err})
		return c.Redirect(http.StatusFound, h.frontendURL+"?err=spotifyAuthFailed")
	}

	state := c.QueryParam("state")
	if state == "" {
		logger.Error("spotify auth failed: state is empty")
		return c.Redirect(http.StatusFound, h.frontendURL+"?err=spotifyAuthFailed")
	}

	code := c.QueryParam("code")
	if code == "" {
		logger.Error("spotify auth failed: code is empty")
		return c.Redirect(http.StatusFound, h.frontendURL+"?err=spotifyAuthFailed")
	}

	redirectURL, sessionID, err := h.authUC.Authorization(state, code)
	if err != nil {
		logger.Errorj(map[string]interface{}{"message": "spotify auth failed", "error": err.Error()})
		if redirectURL == "" {
			return c.Redirect(http.StatusFound, h.frontendURL+"?err=spotifyAuthFailed")
		}
		return c.Redirect(http.StatusFound, redirectURL+"?err=spotifyAuthFailed")
	}

	sameSite := http.SameSiteNoneMode
	if config.IsLocal() {
		sameSite = http.SameSiteLaxMode
	}

	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   sevenDays,
		Secure:   !config.IsLocal(),
		HttpOnly: true,
		SameSite: sameSite,
	})

	return c.Redirect(http.StatusFound, redirectURL)
}
