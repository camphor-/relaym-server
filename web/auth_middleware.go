package web

import (
	"net/http"

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
// TODO テストを書く
func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessCookie, err := c.Cookie("session")
		if err != nil {
			c.Logger().Warn("session cookie not found err=%v", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}
		userID, err := m.uc.GetUserIDFromSession(sessCookie.Value)
		if err != nil {
			c.Logger().Warnf("failed to get session  sessionID=%s err=%v", sessCookie.Value, err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		token, err := m.uc.GetTokenByUserID(userID)
		if err != nil {
			c.Logger().Errorf("failed to get token userID=%s err=%v", userID, err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}
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
