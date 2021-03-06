package ws

import (
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/log"
)

// Hub は すべてのWebSocketクライアント一元管理する構造体です。
// プロセス内に実体は一つしか存在しません。
type Hub struct {
	// １つ目のキーがセッションID
	// O(1) で Client を削除できるようにmapでClientを持つ
	clientsPerSession map[string]map[*Client]struct{}
	pushMsgCh         chan *event.PushMessage
	registerCh        chan *Client
	unregisterCh      chan *Client
}

// NewHub はHubのポインタを生成します。
func NewHub() *Hub {
	return &Hub{
		clientsPerSession: map[string]map[*Client]struct{}{},
		pushMsgCh:         make(chan *event.PushMessage, 10),
		registerCh:        make(chan *Client, 10),
		unregisterCh:      make(chan *Client, 10),
	}
}

// UnregisterCh は送信専用のクライアント登録解除のチャネルを返します。
func (h *Hub) UnregisterCh() chan<- *Client {
	return h.unregisterCh
}

// Register は新しいWebSocketのクライアントをHubに登録します。
// スレッドセーフになるようにチャネルを通じて登録されます。
// 実際の作業は Run() で行われます。
func (h *Hub) Register(client *Client) {
	h.registerCh <- client
}

// Unregister はWebSocketのクライアントをHubから登録解除します。
// スレッドセーフになるようにチャネルを通じて登録解除されます。
// 実際の作業は Run() で行われます。
func (h *Hub) Unregister(client *Client) {
	h.unregisterCh <- client
}

// Push はpushMsgをチャネルに流して、接続されているクライアントに送信します。
// event.Pusher インターフェースを満たしています。
func (h *Hub) Push(pushMsg *event.PushMessage) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{
		"message":   "push message",
		"sessionID": pushMsg.SessionID,
		"event":     pushMsg.Msg,
	})
	h.pushMsgCh <- pushMsg
}

// Run はWebSocketのメッセージを送信するメインループを実行する関数です。
func (h *Hub) Run() {
	for {
		select {
		case cli := <-h.registerCh:
			h.register(cli)
		case cli := <-h.unregisterCh:
			h.unregister(cli)
		case pushMsg := <-h.pushMsgCh:
			h.push(pushMsg)
		}
	}
}

func (h *Hub) register(cli *Client) {
	logger := log.New()
	sessionID := cli.sessionID
	logger.Debugj(map[string]interface{}{"message": "register websocket", "sessionID": sessionID})

	if _, ok := h.clientsPerSession[sessionID]; ok {
		h.clientsPerSession[sessionID][cli] = struct{}{}
		return
	}
	h.clientsPerSession[sessionID] = map[*Client]struct{}{cli: {}}
}

func (h *Hub) unregister(cli *Client) {
	logger := log.New()
	sessionID := cli.sessionID
	logger.Debugj(map[string]interface{}{"message": "unregister websocket", "sessionID": sessionID})

	if _, ok := h.clientsPerSession[sessionID][cli]; ok {
		delete(h.clientsPerSession[sessionID], cli)
		return
	}
}

func (h *Hub) push(pushMsg *event.PushMessage) {
	for cli := range h.clientsPerSession[pushMsg.SessionID] {
		cli.pushCh <- pushMsg.Msg
	}
}
