// +build !race

package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/websocket"
)

func TestHub_UnregisterCh(t *testing.T) {
	ch := make(chan *Client)

	tests := []struct {
		name         string
		unregisterCh chan *Client
		want         chan<- *Client
	}{
		{
			name:         "正しくチャネルを取得できる",
			unregisterCh: ch,
			want:         ch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hub{
				unregisterCh: tt.unregisterCh,
			}
			if got := h.UnregisterCh(); !cmp.Equal(tt.want, got) {
				t.Errorf("UnregisterCh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHub_Register(t *testing.T) {
	newCli := &Client{
		sessionID: "sessionID",
		ws:        &websocket.Conn{},
	}
	sameCli := &Client{
		sessionID: "sessionID",
		ws:        &websocket.Conn{},
	}
	existingCli := &Client{
		sessionID: "existingSessionID",
		ws:        &websocket.Conn{},
	}
	tests := []struct {
		name              string
		clientsPerSession map[string]map[*Client]struct{}
		registerCh        chan *Client
		client            *Client
		want              map[string]map[*Client]struct{}
	}{
		{
			name:              "Clientが0個の時に新規のClientを追加できる",
			clientsPerSession: map[string]map[*Client]struct{}{},
			registerCh:        make(chan *Client),
			client:            newCli,
			want:              map[string]map[*Client]struct{}{"sessionID": {newCli: struct{}{}}},
		},
		{
			name:              "Clientが既に存在する時に同じsessionIDの新規のClientを追加できる",
			clientsPerSession: map[string]map[*Client]struct{}{"sessionID": {sameCli: struct{}{}}},
			registerCh:        make(chan *Client),
			client:            newCli,
			want:              map[string]map[*Client]struct{}{"sessionID": {newCli: struct{}{}, sameCli: struct{}{}}},
		},
		{
			name:              "Clientが既に存在する時に別のsessionIDの新規のClientを追加できる",
			clientsPerSession: map[string]map[*Client]struct{}{"existingSessionID": {existingCli: struct{}{}}},
			registerCh:        make(chan *Client),
			client:            newCli,
			want:              map[string]map[*Client]struct{}{"sessionID": {newCli: struct{}{}}, "existingSessionID": {existingCli: struct{}{}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hub{
				clientsPerSession: tt.clientsPerSession,
				registerCh:        tt.registerCh,
			}
			go h.Run()
			h.Register(tt.client)

			if !cmp.Equal(tt.want, h.clientsPerSession) {
				t.Errorf("Register() diff=%v", cmp.Diff(tt.want, h.clientsPerSession))
			}
		})
	}
}

func TestHub_Unregister(t *testing.T) {
	deletingCli := &Client{
		sessionID: "sessionID",
		ws:        &websocket.Conn{},
	}
	sameCli := &Client{
		sessionID: "sessionID",
		ws:        &websocket.Conn{},
	}
	existingCli := &Client{
		sessionID: "existingSessionID",
		ws:        &websocket.Conn{},
	}

	tests := []struct {
		name              string
		clientsPerSession map[string]map[*Client]struct{}
		unregisterCh      chan *Client
		client            *Client
		want              map[string]map[*Client]struct{}
	}{

		{
			name:              "存在しないClientを削除しても何も起こらない",
			clientsPerSession: map[string]map[*Client]struct{}{"existingSessionID": {existingCli: struct{}{}}},
			unregisterCh:      make(chan *Client),
			client:            deletingCli,
			want:              map[string]map[*Client]struct{}{"existingSessionID": {existingCli: struct{}{}}},
		},
		{
			name:              "存在するClientを削除できる",
			clientsPerSession: map[string]map[*Client]struct{}{"sessionID": {deletingCli: struct{}{}}},
			unregisterCh:      make(chan *Client),
			client:            deletingCli,
			want:              map[string]map[*Client]struct{}{"sessionID": {}},
		},
		{
			name:              "存在するClientを消しても他のClientは削除されない",
			clientsPerSession: map[string]map[*Client]struct{}{"sessionID": {sameCli: struct{}{}, deletingCli: struct{}{}}, "existingSessionID": {existingCli: struct{}{}}},
			unregisterCh:      make(chan *Client),
			client:            deletingCli,
			want:              map[string]map[*Client]struct{}{"sessionID": {sameCli: struct{}{}}, "existingSessionID": {existingCli: struct{}{}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hub{
				clientsPerSession: tt.clientsPerSession,
				unregisterCh:      tt.unregisterCh,
			}

			go h.Run()
			h.Unregister(tt.client)

			if !cmp.Equal(tt.want, h.clientsPerSession) {
				t.Errorf("Unregister() diff=%v", cmp.Diff(tt.want, h.clientsPerSession))
			}
		})
	}
}

func TestHub_Push(t *testing.T) {
	// WebSocketのコネクションを準備
	s := &testWSServer{}
	ts := httptest.NewServer(s)
	url := strings.Replace(ts.URL, "http://", "ws://", 1)

	closedHeader := http.Header{}
	closedHeader.Add("X-CONNECTION-STATUS-FOR-TEST", "Closed")
	closedConn, _, err := websocket.DefaultDialer.Dial(url, closedHeader)
	if err != nil {
		t.Fatal(err)
	}
	if err := closedConn.Close(); err != nil {
		t.Fatal(err)
	}
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}

	cli := NewClient("sessionID", conn, nil)
	invalidCli := NewClient("sessionID", closedConn, nil)
	go cli.PushLoop()
	go invalidCli.PushLoop()

	tests := []struct {
		name              string
		clientsPerSession map[string]map[*Client]struct{}
		unregisterCh      chan *Client
		pushMsgCh         chan *event.PushMessage
		pushMsg           *event.PushMessage
		want              map[string]map[*Client]struct{}
		wantErr           bool
	}{
		{
			name:              "正常に送信されたらClientは削除されない",
			clientsPerSession: map[string]map[*Client]struct{}{"sessionID": {cli: struct{}{}}},
			unregisterCh:      make(chan *Client, 10),
			pushMsgCh:         make(chan *event.PushMessage),
			pushMsg: &event.PushMessage{
				SessionID: "sessionID",
				Msg:       &entity.Event{Type: "ADDTRACK"},
			},
			want:    map[string]map[*Client]struct{}{"sessionID": {cli: struct{}{}}},
			wantErr: false,
		},
		{
			name:              "送信に失敗した時にClientが削除される",
			clientsPerSession: map[string]map[*Client]struct{}{"sessionID": {invalidCli: struct{}{}}},
			unregisterCh:      make(chan *Client, 10),
			pushMsgCh:         make(chan *event.PushMessage),
			pushMsg: &event.PushMessage{
				SessionID: "sessionID",
				Msg:       &entity.Event{Type: "ADDTRACK"},
			},
			want:    map[string]map[*Client]struct{}{"sessionID": {}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hub{
				clientsPerSession: tt.clientsPerSession,
				pushMsgCh:         tt.pushMsgCh,
				registerCh:        make(chan *Client),
				unregisterCh:      tt.unregisterCh,
			}

			cli.notifyClosedCh = h.UnregisterCh()
			invalidCli.notifyClosedCh = h.UnregisterCh()

			go h.Run()
			h.Push(tt.pushMsg)

			time.Sleep(100 * time.Millisecond)
			if !cmp.Equal(tt.want, h.clientsPerSession) {
				t.Errorf("Push() diff=%v", cmp.Diff(tt.want, h.clientsPerSession))
			}

			if !tt.wantErr {
				var gotEvent *entity.Event
				s.ws.SetReadDeadline(time.Now().Add(1 * time.Second))
				if err := s.ws.ReadJSON(&gotEvent); err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(tt.pushMsg.Msg, gotEvent) {
					t.Errorf("Push() recieved message diff=%v", cmp.Diff(tt.pushMsg.Msg, gotEvent))
				}
			}
		})
	}
}

type testWSServer struct {
	ws *websocket.Conn
}

func (s *testWSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, _ := upgrader.Upgrade(w, r, nil)

	if r.Header.Get("X-CONNECTION-STATUS-FOR-TEST") == "Closed" {
		return
	}
	s.ws = ws
}
