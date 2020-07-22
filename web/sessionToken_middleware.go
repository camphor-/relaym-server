package web

import (
	"errors"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/service"
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

		loginUserID := ""
		sessCookie, err := c.Cookie("session")
		if err == nil {
			userID, err := m.uc.GetUserIDFromSession(sessCookie.Value)
			if err == nil {
				loginUserID = userID
			}
		}

		token, creatorID, err := m.uc.GetTokenAndCreatorIDBySessionID(sessionID)
		if err != nil {
			if errors.Is(err, entity.ErrSessionNotFound) {
				logger.Warn(err)
				return echo.NewHTTPError(http.StatusNotFound)
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

		c = setToCreatorContext(c, loginUserID, creatorID, token)
		return next(c)
	}
}
func setToCreatorContext(c echo.Context, userID, creatorID string, token *oauth2.Token) echo.Context {
	ctx := c.Request().Context()
	ctx = service.SetUserIDToContext(ctx, userID)
	ctx = service.SetCreatorIDToContext(ctx, creatorID)
	ctx = service.SetTokenToContext(ctx, token)
	c.SetRequest(c.Request().WithContext(ctx))
	return c
}
