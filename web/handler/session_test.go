package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

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
		name                     string
		sessionID                string
		body                     string
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		wantErr                  bool
		wantCode                 int
	}{
		{
			name:                     "存在しないstateのとき400",
			sessionID:                "sessionID",
			body:                     `{"state": "INVALID"}`,
			prepareMockPlayerFn:      func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:      func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn:    func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {},
			wantErr:                  true,
			wantCode:                 http.StatusBadRequest,
		},
		{
			name:                  "PLAYで指定されたidのセッションが存在しないとき404",
			sessionID:             "notFoundSessionID",
			body:                  `{"state": "PLAY"}`,
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
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
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
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
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.EventPlay,
				})
			},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
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
			name:                  "PAUSEで指定されたidのセッションが存在しないとき404",
			sessionID:             "notFoundSessionID",
			body:                  `{"state": "PAUSE"}`,
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
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
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
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
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
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
			mockUserRepo := mock_repository.NewMockUser(ctrl)
			tt.prepareMockUserRepoFn(mockUserRepo)
			mockSessionRepo := mock_repository.NewMockSession(ctrl)
			tt.prepareMockSessionRepoFn(mockSessionRepo)
			uc := usecase.NewSessionUseCase(mockSessionRepo, mockUserRepo, mockPlayer, mockPusher)
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

func TestSessionHandler_SetDevice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		userID                string
		sessionID             string
		body                  string
		prepareMockRepoFn     func(m *mock_repository.MockSession)
		prepareMockUserRepoFn func(m *mock_repository.MockUser)
		wantErr               bool
		wantCode              int
	}{
		{
			name:                  "デバイスIDが空だと400",
			userID:                "user_id",
			sessionID:             "sessionID",
			body:                  `{"device_id": ""}`,
			prepareMockRepoFn:     func(m *mock_repository.MockSession) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			wantErr:               true,
			wantCode:              http.StatusBadRequest,
		},
		{
			name:      "セッションが存在しないと404",
			userID:    "user_id",
			sessionID: "session_id",
			body:      `{"device_id": "device_id"}`,
			prepareMockRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("session_id").Return(nil, entity.ErrSessionNotFound)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			wantErr:               true,
			wantCode:              http.StatusNotFound,
		},
		{
			name:      "リクエストしたユーザがセッションの作成者ではないと403",
			userID:    "user_id",
			sessionID: "session_id",
			body:      `{"device_id": "device_id"}`,
			prepareMockRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("session_id").Return(&entity.Session{
					ID:          "session_id",
					Name:        "name",
					CreatorID:   "creator_id",
					QueueHead:   0,
					StateType:   "PAUSE",
					QueueTracks: nil,
				}, nil)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			wantErr:               true,
			wantCode:              http.StatusForbidden,
		},
		{
			name:      "正しくデバイスをセットできると204",
			userID:    "creator_id",
			sessionID: "session_id",
			body:      `{"device_id": "device_id"}`,
			prepareMockRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("session_id").Return(&entity.Session{
					ID:          "session_id",
					Name:        "name",
					CreatorID:   "creator_id",
					DeviceID:    "",
					StateType:   "PAUSE",
					QueueHead:   0,
					QueueTracks: nil,
				}, nil)
				m.EXPECT().Update(&entity.Session{
					ID:          "session_id",
					Name:        "name",
					CreatorID:   "creator_id",
					DeviceID:    "device_id",
					StateType:   "PAUSE",
					QueueHead:   0,
					QueueTracks: nil,
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			wantErr:               false,
			wantCode:              http.StatusNoContent,
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
			c.SetPath("/sessions/:id/devices")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)
			c = setToContext(c, tt.userID, nil)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := mock_repository.NewMockSession(ctrl)
			tt.prepareMockRepoFn(mockRepo)
			mockUserRepo := mock_repository.NewMockUser(ctrl)
			tt.prepareMockUserRepoFn(mockUserRepo)

			uc := usecase.NewSessionUseCase(mockRepo, mockUserRepo, nil, nil)
			h := &SessionHandler{uc: uc}

			err := h.SetDevice(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetDevice() error = %v, wantErr %v", err, tt.wantErr)
			}

			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("SetDevice() code = %d, want = %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestSessionHandler_PostSession(t *testing.T) {
	sessionResponse := &sessionRes{
		ID:   "ID",
		Name: "go! go! session!",
		Creator: creatorJSON{
			ID:          "creatorID",
			DisplayName: "creatorDisplayName",
		},
		Playback: playbackJSON{
			State: stateJSON{
				Type: "STOP",
			},
			Device: nil,
		},
		Queue: queueJSON{
			Head:   0,
			Tracks: nil,
		},
	}
	user := &entity.User{
		ID:            "creatorID",
		SpotifyUserID: "creatorSpotifyUserID",
		DisplayName:   "creatorDisplayName",
	}
	tests := []struct {
		name                     string
		body                     string
		userID                   string
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		want                     *sessionRes
		wantErr                  bool
		wantCode                 int
	}{
		{
			name:                "nameを渡すと正常に動作する",
			body:                `{"name": "go! go! session!"}`,
			userID:              "creatorID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().StoreSession(gomock.Any()).Return(nil)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {
				m.EXPECT().FindByID("creatorID").Return(user, nil)
			},
			want:     sessionResponse,
			wantErr:  false,
			wantCode: http.StatusCreated,
		},
		{
			name:                     "nameが空だとempty nameが返る",
			body:                     `{"name": ""}`,
			userID:                   "creatorID",
			prepareMockPlayerFn:      func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:      func(m *mock_event.MockPusher) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {},
			prepareMockUserRepoFn:    func(m *mock_repository.MockUser) {},
			want:                     sessionResponse,
			wantErr:                  true,
			wantCode:                 http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c = setToContext(c, tt.userID, nil)
			c.SetPath("/sessions")

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockPlayer := mock_spotify.NewMockPlayer(ctrl)
			tt.prepareMockPlayerFn(mockPlayer)
			mockPusher := mock_event.NewMockPusher(ctrl)
			tt.prepareMockPusherFn(mockPusher)
			mockSessionRepo := mock_repository.NewMockSession(ctrl)
			tt.prepareMockSessionRepoFn(mockSessionRepo)
			mockUserRepo := mock_repository.NewMockUser(ctrl)
			tt.prepareMockUserRepoFn(mockUserRepo)
			uc := usecase.NewSessionUseCase(mockSessionRepo, mockUserRepo, mockPlayer, mockPusher)
			h := &SessionHandler{
				uc: uc,
			}
			postErr := h.PostSession(c)
			if (postErr != nil) != tt.wantErr {
				t.Errorf("PostSession() error = %v, wantErr %v", postErr, tt.wantErr)
			}

			// ステータスコードのチェック
			if er, ok := postErr.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("PostSession() code = %d, want = %d", rec.Code, tt.wantCode)
			}

			if !tt.wantErr {
				got := &sessionRes{}
				err := json.Unmarshal(rec.Body.Bytes(), got)
				if err != nil {
					t.Fatal(err)
				}
				opts := []cmp.Option{cmpopts.IgnoreFields(sessionRes{}, "ID")}
				if !cmp.Equal(got, tt.want, opts...) {
					t.Errorf("PostSession() diff = %v", cmp.Diff(got, tt.want, opts...))
				}
			}
		})
	}
}

func TestSessionHandler_AddQueue(t *testing.T) {
	session := &entity.Session{
		ID:        "sessionID",
		Name:      "sessionName",
		CreatorID: "sessionCreator",
		DeviceID:  "sessionDeviceID",
		StateType: "PLAY",
		QueueHead: 0,
		QueueTracks: []*entity.QueueTrack{{
			Index:     0,
			URI:       "spotify:track:existed_session_uri",
			SessionID: "sessionID",
		},
		},
	}

	tests := []struct {
		name                     string
		sessionID                string
		body                     string
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		wantErr                  bool
		wantCode                 int
	}{
		{
			name:      "正しいuriが渡されると正常に動作する",
			sessionID: "sessionID",
			body:      `{"uri": "spotify:track:valid_uri"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().AddToQueue(gomock.Any(), "spotify:track:valid_uri", "sessionDeviceID").Return(nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.EventAddTrack,
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(session, nil)
				m.EXPECT().StoreQueueTrack(&entity.QueueTrackToStore{
					URI:       "spotify:track:valid_uri",
					SessionID: "sessionID",
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusNoContent,
		},
		{
			name:                     "uriが空の時400",
			sessionID:                "sessionID",
			body:                     `{"uri": ""}`,
			prepareMockPlayerFn:      func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:      func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn:    func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {},
			wantErr:                  true,
			wantCode:                 http.StatusBadRequest,
		},
		{
			name:                  "存在しないsessionIDの時404",
			sessionID:             "invalidSessionID",
			body:                  `{"uri": "valid_uri"}`,
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("invalidSessionID").Return(nil, entity.ErrSessionNotFound)
			},
			wantErr:  true,
			wantCode: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/sessions/:id/queue")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockPlayer := mock_spotify.NewMockPlayer(ctrl)
			tt.prepareMockPlayerFn(mockPlayer)
			mockPusher := mock_event.NewMockPusher(ctrl)
			tt.prepareMockPusherFn(mockPusher)
			mockUserRepo := mock_repository.NewMockUser(ctrl)
			tt.prepareMockUserRepoFn(mockUserRepo)
			mockSessionRepo := mock_repository.NewMockSession(ctrl)
			tt.prepareMockSessionRepoFn(mockSessionRepo)
			uc := usecase.NewSessionUseCase(mockSessionRepo, mockUserRepo, mockPlayer, mockPusher)
			h := &SessionHandler{
				uc: uc,
			}
			err := h.AddQueue(c)
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
