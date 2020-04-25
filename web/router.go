package web

import (
	"github.com/camphor-/relaym-server/usecase"
	"github.com/camphor-/relaym-server/web/handler"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewServer はミドルウェアやハンドラーが登録されたechoの構造体を返します。
func NewServer(authUC *usecase.AuthUseCase, userUC *usecase.UserUseCase) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CSRF())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"relaym.local:3000"}, // TODO : 環境変数から読み込むようにする
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	userHandler := handler.NewUserHandler(userUC)

	// TODO フロントエンドのURLを環境変数で指定する
	authHandler := handler.NewAuthHandler(authUC, "http://relaym.local:3000")

	v3 := e.Group("/api/v3")
	v3.GET("/login", authHandler.Login)
	v3.GET("/callback", authHandler.Callback)

	user := v3.Group("/users", NewAuthMiddleware(authUC).Authenticate)
	user.GET("/me", userHandler.GetMe)
	return e
}
