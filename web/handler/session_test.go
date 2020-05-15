package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/mock_event"
	"github.com/camphor-/relaym-server/domain/mock_spotify"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
)

func TestSessionHandler_Playback(t *testing.T) {
	tests := []struct {
		name                string
		sessionID           string
		body                string
		prepareMockPlayerFn func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn func(m *mock_event.MockPusher)
		wantErr             bool
		wantCode            int
	}{
		{
			name:                "存在しないstateのとき400",
			sessionID:           "sessionID",
			body:                `{"state": "INVALID"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			wantErr:             true,
			wantCode:            http.StatusBadRequest,
		},
		// TODO レポジトリが実装されたら有効にする
		// {
		// 	name:                "PLAYで指定されたidのセッションが存在しないとき404",
		// 	sessionID:           "notFoundSessionID",
		// 	body:                `{"state": "PLAY"}`,
		// 	prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
		// 	prepareMockPusherFn: func(m *mock_event.MockPusher) {},
		// 	wantErr:             true,
		// 	wantCode:            http.StatusNotFound,
		// },
		{
			name:      "PLAYで再生するデバイスがオフラインのとき403",
			sessionID: "sessionID",
			body:      `{"state": "PLAY"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Play(gomock.Any(), "").Return(entity.ErrActiveDeviceNotFound)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			wantErr:             true,
			wantCode:            http.StatusForbidden,
		},
		{
			name:      "PLAYで正しく再生リクエストが処理されたとき202",
			sessionID: "sessionID",
			body:      `{"state": "PLAY"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Play(gomock.Any(), "").Return(nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.EventPlay,
				})
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/sessions/:id/playback")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockPlayer := mock_spotify.NewMockPlayer(ctrl)
			tt.prepareMockPlayerFn(mockPlayer)
			mockPusher := mock_event.NewMockPusher(ctrl)
			tt.prepareMockPusherFn(mockPusher)
			uc := usecase.NewSessionUseCase(mockPlayer, mockPusher)
			h := &SessionHandler{
				uc: uc,
			}
			err := h.Playback(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Playback() error = %v, wantErr %v", err, tt.wantErr)
			}

			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("Playback() code = %d, want = %d", rec.Code, tt.wantCode)
			}
		})
	}
}
