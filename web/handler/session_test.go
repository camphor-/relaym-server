package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/mock_event"
	"github.com/camphor-/relaym-server/domain/mock_repository"
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
		prepareMockRepoFn   func(m *mock_repository.MockSession)
		wantErr             bool
		wantCode            int
	}{
		{
			name:                "存在しないstateのとき400",
			sessionID:           "sessionID",
			body:                `{"state": "INVALID"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			prepareMockRepoFn:   func(m *mock_repository.MockSession) {},
			wantErr:             true,
			wantCode:            http.StatusBadRequest,
		},
		{
			name:                "PLAYで指定されたidのセッションが存在しないとき404",
			sessionID:           "notFoundSessionID",
			body:                `{"state": "PLAY"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			prepareMockRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("notFoundSessionID").Return(nil, entity.ErrSessionNotFound)
			},
			wantErr:  true,
			wantCode: http.StatusNotFound,
		},
		{
			name:      "PLAYで再生するデバイスがオフラインのとき403",
			sessionID: "sessionID",
			body:      `{"state": "PLAY"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Play(gomock.Any(), "").Return(entity.ErrActiveDeviceNotFound)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			prepareMockRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:          "sessionID",
					Name:        "session_name",
					CreatorID:   "creator_id",
					QueueHead:   0,
					StateType:   "PLAY",
					QueueTracks: nil,
				}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusForbidden,
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
			prepareMockRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:          "sessionID",
					Name:        "session_name",
					CreatorID:   "creator_id",
					QueueHead:   0,
					StateType:   "STOP",
					QueueTracks: nil,
				}, nil)
				m.EXPECT().Update(&entity.Session{
					ID:          "sessionID",
					Name:        "session_name",
					CreatorID:   "creator_id",
					QueueHead:   0,
					StateType:   "PLAY",
					QueueTracks: nil,
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:                "PAUSEで指定されたidのセッションが存在しないとき404",
			sessionID:           "notFoundSessionID",
			body:                `{"state": "PAUSE"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			prepareMockRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("notFoundSessionID").Return(nil, entity.ErrSessionNotFound)
			},
			wantErr:  true,
			wantCode: http.StatusNotFound,
		},
		{
			name:      "PAUSEで再生するデバイスがオフラインのときは、既に再生が止まっているはずなので202を返す",
			sessionID: "sessionID",
			body:      `{"state": "PAUSE"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Pause(gomock.Any(), "").Return(entity.ErrActiveDeviceNotFound)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			prepareMockRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:          "sessionID",
					Name:        "session_name",
					CreatorID:   "creator_id",
					QueueHead:   0,
					StateType:   "PAUSE",
					QueueTracks: nil,
				}, nil)
				m.EXPECT().Update(&entity.Session{
					ID:          "sessionID",
					Name:        "session_name",
					CreatorID:   "creator_id",
					QueueHead:   0,
					StateType:   "PAUSE",
					QueueTracks: nil,
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:      "PAUSEで正しく再生リクエストが処理されたとき202",
			sessionID: "sessionID",
			body:      `{"state": "PAUSE"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Pause(gomock.Any(), "").Return(nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
			},
			prepareMockRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:          "sessionID",
					Name:        "session_name",
					CreatorID:   "creator_id",
					QueueHead:   0,
					StateType:   "PLAY",
					QueueTracks: nil,
				}, nil)
				m.EXPECT().Update(&entity.Session{
					ID:          "sessionID",
					Name:        "session_name",
					CreatorID:   "creator_id",
					QueueHead:   0,
					StateType:   "PAUSE",
					QueueTracks: nil,
				}).Return(nil)
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
			mockRepo := mock_repository.NewMockSession(ctrl)
			tt.prepareMockRepoFn(mockRepo)
			uc := usecase.NewSessionUseCase(mockRepo, mockPlayer, mockPusher)
			h := &SessionHandler{
				uc: uc,
			}
			err := h.Playback(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Playback() error = %v, wantErr %v", err, tt.wantErr)
			}

			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); ok && er.Code != tt.wantCode {
				t.Errorf("Playback() code = %d, want = %d", rec.Code, tt.wantCode)
			}
		})
	}
}
