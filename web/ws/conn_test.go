package ws

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gorilla/websocket"
)

func TestClient_PingLoop(t *testing.T) {
	// WebSocketのコネクションを準備
	s := &testWSServer{}
	ts := httptest.NewServer(s)
	url := strings.Replace(ts.URL, "http://", "ws://", 1)

	closedConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := closedConn.Close(); err != nil {
		t.Fatal(err)
	}

	tmpPingWait := pingPeriod
	pingPeriod = 1 * time.Second
	defer func() {
		pingPeriod = tmpPingWait
	}()

	ch := make(chan *Client)

	tests := []struct {
		name           string
		sessionID      string
		ws             *websocket.Conn
		notifyClosedCh chan<- *Client
	}{
		{
			name:           "pingに失敗したときにnotifyClosedChにClientが飛ぶ",
			sessionID:      "sessionID",
			ws:             closedConn,
			notifyClosedCh: ch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				sessionID:      tt.sessionID,
				ws:             tt.ws,
				notifyClosedCh: tt.notifyClosedCh,
			}
			go c.PingLoop()

			cli := <-ch

			opts := []cmp.Option{cmp.AllowUnexported(Client{}), cmpopts.IgnoreUnexported(websocket.Conn{})}
			if !cmp.Equal(c, cli, opts...) {
				t.Errorf("PingLoop() notifyClosedCh diff=%v", cmp.Diff(c, ch, opts...))

			}

		})
	}
}
