package web

import (
	"errors"
	"net/http"

	"github.com/camphor-/relaym-server/log"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// AuthMiddleware は認証を担当するミドルウェアを管理する構造体です。
type AuthMiddleware struct {
	uc *usecase.AuthUseCase
}

// NewAuthMiddleware web.AuthMiddlewareのポインタを生成します。
func NewAuthMiddleware(uc *usecase.AuthUseCase) *AuthMiddleware {
	return &AuthMiddleware{uc: uc}
}

// Authenticate は認証が必要なAPIで認証情報があるかチェックします。
func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	logger := log.New()
	return func(c echo.Context) error {
		sessCookie, err := c.Cookie("session")
		if err != nil {
			logger.Warnj(map[string]interface{}{"message": "session cookie not found", "error": err.Error()})
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		userID, err := m.uc.GetUserIDFromSession(sessCookie.Value)
		if err != nil {
			logger.Warnj(map[string]interface{}{"message": "failed to get session", "sessionID": sessCookie.Value, "error": err.Error()})
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		token, err := m.uc.GetTokenByUserID(userID)
		if err != nil {
			if errors.Is(err, entity.ErrTokenNotFound) {
				logger.Debug(err)
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			logger.Errorj(map[string]interface{}{"message": "failed to get token", "userID": userID, "sessionID": sessCookie.Value, "error": err.Error()})
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		newToken, err := m.uc.RefreshAccessToken(userID, token)
		if err != nil {
			logger.Errorj(map[string]interface{}{"message": "failed to refresh access token", "sessionID": sessCookie.Value, "error": err.Error()})
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		token = newToken

		c = setToContext(c, userID, token)
		return next(c)
	}
}

func setToContext(c echo.Context, userID string, token *oauth2.Token) echo.Context {
	ctx := c.Request().Context()
	ctx = service.SetUserIDToContext(ctx, userID)
	ctx = service.SetTokenToContext(ctx, token)
	c.SetRequest(c.Request().WithContext(ctx))
	return c
}
