package web

import (
	"github.com/camphor-/relaym-server/database"
	"github.com/camphor-/relaym-server/usecase"
	"github.com/camphor-/relaym-server/web/handler"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewRouter はミドルウェアやハンドラーが登録されたechoのルータを返します。
func NewRouter() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CSRF())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"localhost.local:3000"}, // TODO : 環境変数から読み込むようにする
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	userHandler := handler.NewUserHandler(usecase.NewUserUseCase(database.NewUserRepository()))

	v3 := e.Group("/api/v3")
	user := v3.Group("/users")
	user.GET("/me", userHandler.GetMe)
	return e
}
