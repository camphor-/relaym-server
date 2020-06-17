package web

import (
	"errors"
	"net/http"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/log"
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
	logger := log.New()
	return func(c echo.Context) error {
		sessionID := c.Param("id")
		if sessionID == "" {
			logger.Error("sessionID not found")
			return echo.NewHTTPError(http.StatusNotFound)
		}

		token, creatorID, err := m.uc.GetTokenAndCreatorIDBySessionID(sessionID)
		if err != nil {
			if errors.Is(err, entity.ErrTokenNotFound) {
				logger.Warn(err)
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			logger.Errorj(map[string]interface{}{"message": "failed to get token", "sessionID": sessionID, "error": err.Error()})
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		newToken, err := m.uc.RefreshAccessToken(creatorID, token)
		if err != nil {
			logger.Errorj(map[string]interface{}{"message": "failed to refresh access token", "sessionID": sessionID, "error": err.Error()})
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		token = newToken

		c = setToContext(c, creatorID, token)
		return next(c)
	}
}
