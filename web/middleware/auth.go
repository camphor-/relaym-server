package middleware

import (
	"net/http"

	"github.com/camphor-/relaym-server/domain/service"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/labstack/echo/v4"
)

// Auth は認証を担当するミドルウェアを管理する構造体です。
type Auth struct {
	uc *usecase.AuthUseCase
}

// NewAuth はmiddleware.Authのポインタを生成します。
func NewAuth(uc *usecase.AuthUseCase) *Auth {
	return &Auth{uc: uc}
}

// Authenticate は認証が必要なAPIで認証情報があるかチェックします。
// TODO テストを書く
func (m *Auth) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO ユーザIDをセッションから取得する
		userID := "userID"
		token, err := m.uc.GetTokenByUserID(userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}
		ctx := c.Request().Context()
		ctx = service.SetUserIDToContext(ctx, userID)
		ctx = service.SetTokenToContext(ctx, token)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
