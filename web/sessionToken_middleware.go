package web

import (
	"errors"
	"net/http"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/labstack/echo/v4"
)

// CreatorTokenMiddlewareはSessionのCreatorがもつAccessTokenの管理を担当するミドルウェアを管理する構造体です。
type CreatorTokenMiddleware struct {
	uc *usecase.AuthUseCase
}

// NewCreatorTokenMiddleware web.CreatorTokenMiddlewareのポインタを生成します。
func NewCreatorTokenMiddleware(uc *usecase.AuthUseCase) *CreatorTokenMiddleware {
	return &CreatorTokenMiddleware{uc: uc}
}

// SetCreatorTokenToContext はSessionIDからSessionのCreatorがもつAccessTokenをContextにセットします
func (m *CreatorTokenMiddleware) SetCreatorTokenToContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionID := c.Param("id")
		if sessionID == "" {
			c.Logger().Errorf("sessionID not found")
			return echo.NewHTTPError(http.StatusNotFound)
		}

		token, creatorID, err := m.uc.GetTokenAndCreatorIDBySessionID(sessionID)
		if err != nil {
			if errors.Is(err, entity.ErrTokenNotFound) {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			c.Logger().Errorf("failed to get token sessionID=%s err=%v", sessionID, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		newToken, err := m.uc.RefreshAccessToken(creatorID, token)
		if err != nil {
			c.Logger().Errorf("failed to refresh access token: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		token = newToken

		c = setToContext(c, creatorID, token)
		return next(c)
	}
}
