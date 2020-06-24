package web

import (
	"github.com/camphor-/relaym-server/config"
	"github.com/camphor-/relaym-server/usecase"
	"github.com/camphor-/relaym-server/web/handler"
	"github.com/camphor-/relaym-server/web/ws"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewServer はミドルウェアやハンドラーが登録されたechoの構造体を返します。
func NewServer(authUC *usecase.AuthUseCase, userUC *usecase.UserUseCase, sessionUC *usecase.SessionUseCase, trackUC *usecase.TrackUseCase, hub *ws.Hub) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		Skipper: func(c echo.Context) bool {
			token := c.Request().Header.Get("X-CSRF-Token")
			return token == "relaym"
		},
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{config.CORSAllowOrigin()},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	userHandler := handler.NewUserHandler(userUC)
	trackHandler := handler.NewTrackHandler(trackUC)
	sessionHandler := handler.NewSessionHandler(sessionUC)
	authHandler := handler.NewAuthHandler(authUC, config.FrontendURL())
	wsHandler := handler.NewWebSocketHandler(hub, sessionUC)

	v3 := e.Group("/api/v3")
	v3.GET("/login", authHandler.Login)
	v3.GET("/callback", authHandler.Callback)

	authed := v3.Group("", NewAuthMiddleware(authUC).Authenticate)

	user := authed.Group("/users")
	user.GET("/me", userHandler.GetMe)
	user.GET("/me/devices", userHandler.GetActiveDevices)

	authedSession := authed.Group("/sessions")
	authedSession.POST("", sessionHandler.PostSession)
	authedSession.PUT("/:id/devices", sessionHandler.SetDevice)
	authedSession.POST("/:id/queue", sessionHandler.AddQueue)

	SessionWithCreatorToken := v3.Group("/sessions/:id", NewCreatorTokenMiddleware(authUC).SetCreatorTokenToContext)
	SessionWithCreatorToken.GET("", sessionHandler.GetSession)
	SessionWithCreatorToken.GET("/search", trackHandler.SearchTracks)
	SessionWithCreatorToken.PUT("/playback", sessionHandler.Playback)
	SessionWithCreatorToken.GET("/ws", wsHandler.WebSocket)
	return e
}
