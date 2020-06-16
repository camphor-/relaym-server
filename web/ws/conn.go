package ws

import (
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Client はWebSocketのクライアントを表します。
type Client struct {
	sessionID      string
	ws             *websocket.Conn
	pushCh         chan *entity.Event
	notifyClosedCh chan<- *Client // HubのunregisterChをもらう
}

// NewClient は Clientのポインタを生成します。
func NewClient(sessionID string, ws *websocket.Conn, notifyClosedCh chan<- *Client) *Client {
	return &Client{
		sessionID:      sessionID,
		ws:             ws,
		pushCh:         make(chan *entity.Event, 256),
		notifyClosedCh: notifyClosedCh,
	}
}

// ReadLoop はクライアントからのメッセージを受け取るループです。
// 今回はサーバからイベントを送信するのみですが、Pingのやりとりに必要なのでループを回してます。
func (c *Client) ReadLoop() {
	defer func() {
		c.notifyClosedCh <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("readMessage: unexpected error: %v", err)
			}
			break
		}
	}
}

// PushLoop は一つのWebSocketコネクションに対してメッセージを送信するループです。
// 一つのWebSocketコネクションに対して一つのgoroutineでPushLoop()が実行されます。
// 接続が切れた場合はnotifyClosedChを通じてHubに登録されているwsConnを削除してメモリリークを防ぎます。
func (c *Client) PushLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.notifyClosedCh <- c
		c.ws.Close()
	}()
	for {
		select {
		case msg, ok := <-c.pushCh:
			_ = c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				if err := c.ws.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					fmt.Printf("failed to write close message: sessionID=%s: %v\n", c.sessionID, err)
					return
				}
			}

			if err := c.ws.WriteJSON(msg); err != nil {
				fmt.Printf("failed to WriteJSON: %v\n", err)
				return
			}
		case <-ticker.C:
			_ = c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("failed to ping: sessionID=%s: %v\n", c.sessionID, err)
				return
			}
		}
	}
}
