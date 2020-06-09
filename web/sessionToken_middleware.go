package web

import (
	"errors"
	"net/http"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// SessionTokenMiddlewareはSessionのもつTokenの管理を担当するミドルウェアを管理する構造体です。
type SessionTokenMiddleware struct {
	uc *usecase.SessionTokenUseCase
}

// NewSessionTokenMiddleware web.SessionTokenMiddlewareのポインタを生成します。
func NewSessionTokenMiddleware(uc *usecase.SessionTokenUseCase) *SessionTokenMiddleware {
	return &SessionTokenMiddleware{uc: uc}
}

// SetTokenToContext はSessionIDからSessionのもつTokenをContextにセットします
func (m *SessionTokenMiddleware) SetTokenToContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionID := c.Param("id")
		if sessionID == "" {
			c.Logger().Warnf("sessionID not found err=%v", err)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		token, err := m.uc.GetTokenBySessionID(sessionID)
		if err != nil {
			if errors.Is(err, entity.ErrTokenNotFound) {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			c.Logger().Errorf("failed to get token sessionID=%s err=%v", userID, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		newToken, err := m.uc.RefreshAccessToken(sessionID, token)
		if err != nil {
			c.Logger().Errorf("failed to refresh access token: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		token = newToken

		c = setTokenToContext(c, token)
		return next(c)
	}
}

func setTokenToContext(c echo.Context, token *oauth2.Token) echo.Context {
	ctx := c.Request().Context()
	ctx = service.SetTokenToContext(ctx, token)
	c.SetRequest(c.Request().WithContext(ctx))
	return c
}
