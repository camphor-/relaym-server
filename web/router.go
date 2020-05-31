package web

import (
	"github.com/camphor-/relaym-server/config"
	"github.com/camphor-/relaym-server/usecase"
	"github.com/camphor-/relaym-server/web/handler"
	"github.com/camphor-/relaym-server/web/ws"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

// NewServer はミドルウェアやハンドラーが登録されたechoの構造体を返します。
func NewServer(authUC *usecase.AuthUseCase, userUC *usecase.UserUseCase, sessionUC *usecase.SessionUseCase, trackUC *usecase.TrackUseCase, hub *ws.Hub) *echo.Echo {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	if config.IsLocal() {
		e.Logger.SetLevel(log.DEBUG)
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		Skipper: func(c echo.Context) bool {
			token := c.Request().Header.Get("X-CSRF-Token")
			return token != "relaym"
		},
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"relaym.local:3000"}, // TODO : 環境変数から読み込むようにする
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	userHandler := handler.NewUserHandler(userUC)
	trackHandler := handler.NewTrackHandler(trackUC)
	sessionHandler := handler.NewSessionHandler(sessionUC)

	// TODO フロントエンドのURLを環境変数で指定する
	authHandler := handler.NewAuthHandler(authUC, "http://relaym.local:3000")

	wsHandler := handler.NewWebSocketHandler(hub)

	v3 := e.Group("/api/v3")
	v3.GET("/login", authHandler.Login)
	v3.GET("/callback", authHandler.Callback)

	// TODO 本来は認証が必要だがテストのために認証を外しておく
	v3.GET("/ws/:id", wsHandler.WebSocket)

	authed := v3.Group("", NewAuthMiddleware(authUC).Authenticate)
	authed.GET("/search", trackHandler.SearchTracks)

	user := authed.Group("/users")
	user.GET("/me", userHandler.GetMe)
	user.GET("/me/devices", userHandler.GetActiveDevices)

	authedSession := authed.Group("/sessions")
	authedSession.POST("", sessionHandler.PostSession)
	authedSession.PUT("/:id/devices", sessionHandler.SetDevice)
	authedSession.POST("/:id/queue", sessionHandler.AddQueue)

	noAuthedSession := v3.Group("/sessions")
	noAuthedSession.GET("/:id", sessionHandler.GetSession)
	noAuthedSession.PUT("/:id/playback", sessionHandler.Playback)
	return e
}
