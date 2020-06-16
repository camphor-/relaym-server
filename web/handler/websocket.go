package handler

import (
	"errors"
	"net/http"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/usecase"
	"github.com/camphor-/relaym-server/web/ws"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// WebSocketHandler は /ws 以下のエンドポイントを管理する構造体です。
type WebSocketHandler struct {
	hub      *ws.Hub
	upgrader websocket.Upgrader
	uc       *usecase.SessionUseCase
}

// NewWebSocketHandler はWebSocketHandlerのポインタを生成する関数です。
func NewWebSocketHandler(hub *ws.Hub, uc *usecase.SessionUseCase) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// TODO クライアントが準備できたタイミングで適切にセットする
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		uc: uc,
	}
}

// WebSocket は GET /ws/:id に対応するハンドラーです。
func (h *WebSocketHandler) WebSocket(c echo.Context) error {
	sessionID := c.Param("id")

	ctx := c.Request().Context()

	if err := h.uc.CanConnectToPusher(ctx, sessionID); err != nil {
		if errors.Is(err, entity.ErrSessionNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, entity.ErrSessionNotFound.Error())
		}
		c.Logger().Errorf("CanConnectToPusher: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	wsConn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Errorf("upgrader.Upgrade: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	wsCli := ws.NewClient(sessionID, wsConn, h.hub.UnregisterCh())
	h.hub.Register(wsCli)

	go wsCli.PushLoop()
	go wsCli.ReadLoop()

	return nil
}
