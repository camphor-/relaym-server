package handler

import (
	"net/http"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/web/ws"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// WebSocketHandler は /ws 以下のエンドポイントを管理する構造体です。
type WebSocketHandler struct {
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

// NewWebSocketHandler はWebSocketHandlerのポインタを生成する関数です。
func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			// TODO クライアントが準備できたタイミングで適切にセットする
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}}
}

// WebSocket は GET /ws/:id に対応するハンドラーです。
func (h *WebSocketHandler) WebSocket(c echo.Context) error {
	sessionID := c.Param("id")

	wsConn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Errorf("upgrader.Upgrade: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	wsCli := ws.NewClient(sessionID, wsConn, h.hub.UnregisterCh())
	h.hub.Register(wsCli)

	go wsCli.PushLoop()

	// TODO テスト用に置いとくだけで後で消す
	h.hub.Push(&event.PushMessage{
		SessionID: "sessionID",
		Msg: &entity.Event{
			Type: "CONNECTED",
		},
	})

	return nil
}
