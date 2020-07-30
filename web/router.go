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
func NewServer(authUC *usecase.AuthUseCase, userUC *usecase.UserUseCase, sessionUC *usecase.SessionUseCase, sessionStateUC *usecase.SessionStateUseCase, trackUC *usecase.TrackUseCase, batchUC *usecase.BatchUseCase, hub *ws.Hub) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		Skipper: func(c echo.Context) bool {
			token := c.Request().Header.Get("X-CSRF-Token")
			return token == "relaym"
		},
	}))

	allowHeaders := []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "X-CSRF-Token"}
	previewCorsMiddleware := newDeployPreviewCorsMiddleware(allowHeaders, true)

	// `middleware.CORSWithConfig`はOPTIONのときにすぐreturnしてしまい、previewCorsMiddleware.addAllowOriginまで到達しないので
	// ここのミドルウェアでoriginを付与して204を返してしまうようにする
	if config.IsDev() {
		e.Use(previewCorsMiddleware.addAllowOriginForOption)
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{config.CORSAllowOrigin()},
		AllowHeaders:     allowHeaders,
		AllowCredentials: true,
	}))

	if config.IsDev() {
		e.Use(previewCorsMiddleware.addAllowOrigin)
	}

	userHandler := handler.NewUserHandler(userUC)
	trackHandler := handler.NewTrackHandler(trackUC)
	sessionHandler := handler.NewSessionHandler(sessionUC, sessionStateUC)
	authHandler := handler.NewAuthHandler(authUC, config.FrontendURL())
	wsHandler := handler.NewWebSocketHandler(hub, sessionUC)
	batchHandler := handler.NewBatchHandler(batchUC)

	v3 := e.Group("/api/v3")
	v3.GET("/login", authHandler.Login)
	v3.GET("/callback", authHandler.Callback)

	batch := v3.Group("/batch")
	batch.POST("/archive", batchHandler.PostArchive)

	authed := v3.Group("", NewAuthMiddleware(authUC).Authenticate)

	user := authed.Group("/users")
	user.GET("/me", userHandler.GetMe)

	authedSession := authed.Group("/sessions")
	authedSession.POST("", sessionHandler.PostSession)

	sessionWithCreatorToken := v3.Group("/sessions/:id", NewCreatorTokenMiddleware(authUC).SetCreatorTokenToContext)
	sessionWithCreatorToken.GET("", sessionHandler.GetSession)
	sessionWithCreatorToken.GET("/search", trackHandler.SearchTracks)
	sessionWithCreatorToken.GET("/devices", sessionHandler.GetActiveDevices)
	sessionWithCreatorToken.PUT("/devices", sessionHandler.SetDevice)
	sessionWithCreatorToken.POST("/queue", sessionHandler.Enqueue)
	sessionWithCreatorToken.PUT("/state", sessionHandler.State)
	//TODO: PUTに戻す
	sessionWithCreatorToken.GET("/next", sessionHandler.NextTrack)
	sessionWithCreatorToken.GET("/ws", wsHandler.WebSocket)
	return e
}
