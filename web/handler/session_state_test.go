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

func TestSessionHandler_State(t *testing.T) {
	tests := []struct {
		name                     string
		sessionID                string
		body                     string
		userID                   string
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
			userID:                   "userID",
			prepareMockPlayerFn:      func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:      func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn:    func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {},
			wantErr:                  true,
			wantCode:                 http.StatusBadRequest,
		},
		{
			name:                  "作成者以外のリクエストで、他人による操作が許可されていないときは400",
			sessionID:             "sessionID",
			body:                  `{"state": "PLAY"}`,
			userID:                "userID",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Pause,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: false,
				}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:      "作成者以外のリクエストでも他人による操作が許可されているときは202",
			sessionID: "sessionID",
			body:      `{"state": "PLAY"}`,
			userID:    "nonCreatorID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().SetRepeatMode(gomock.Any(), false, "device_id").Return(nil)
				m.EXPECT().SetShuffleMode(gomock.Any(), false, "device_id").Return(nil)
				m.EXPECT().Play(gomock.Any(), "device_id").Return(nil)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{SessionID: "sessionID", Msg: entity.EventPlay})
			},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Pause,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Play,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
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
			c.SetPath("/sessions/:id/state")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)
			c = setToContext(c, tt.userID, nil)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			h := newSessionStateHandlerForTest(t, ctrl, tt.prepareMockPlayerFn, tt.prepareMockPusherFn,
				tt.prepareMockUserRepoFn, tt.prepareMockSessionRepoFn)

			err := h.State(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("State() error = %v, wantErr %v", err, tt.wantErr)
			}

			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); ok && er.Code != tt.wantCode {
				t.Errorf("State() code = %d, want = %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestSessionHandler_State_PLAY(t *testing.T) {
	tests := []struct {
		name                     string
		sessionID                string
		userID                   string
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		wantErr                  bool
		wantCode                 int
	}{
		{
			name:      "StateType=Pause: 再生をし始じめるがSpotifyのキューは変更せず202",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().SetRepeatMode(gomock.Any(), false, "device_id").Return(nil)
				m.EXPECT().SetShuffleMode(gomock.Any(), false, "device_id").Return(nil)
				m.EXPECT().Play(gomock.Any(), "device_id").Return(nil)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{SessionID: "sessionID", Msg: entity.EventPlay})
			},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Pause,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Play,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:      "StateType=PAUSE: 再生するデバイスがオフラインかつセッション作成者以外のリクエストのとき403",
			sessionID: "sessionID",
			userID:    "nonCreatorID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().SetRepeatMode(gomock.Any(), false, "device_id").Return(entity.ErrActiveDeviceNotFound)
			},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Pause,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusForbidden,
		},
		{
			name:      "StateType=STOP: Spotifyのキューをすべてスキップした後、最初の1曲を再生し初めて、2,3曲目のみをSpotifyのqueueに追加し、202",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().SetRepeatMode(gomock.Any(), false, "device_id").Return(nil)
				m.EXPECT().SetShuffleMode(gomock.Any(), false, "device_id").Return(nil)
				m.EXPECT().SkipAllTracks(gomock.Any(), "device_id", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF").Return(nil)
				m.EXPECT().PlayWithTracks(gomock.Any(), "device_id", []string{"spotify:track:5uQ0vKy2973Y9IUCd1wMEF"}).Return(nil)
				m.EXPECT().Enqueue(gomock.Any(), "spotify:track:49BRCNV7E94s7Q2FUhhT3w", "device_id").Return(nil)
				m.EXPECT().Enqueue(gomock.Any(), "spotify:track:3", "device_id").Return(nil)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.EventPlay,
				})
			},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Stop,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
						{Index: 2, URI: "spotify:track:3"},
						{Index: 2, URI: "spotify:track:4"},
					},
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: "PLAY",
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
						{Index: 2, URI: "spotify:track:3"},
						{Index: 2, URI: "spotify:track:4"},
					},
					AllowToControlByOthers: true,
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:      "StateType=STOP: キューに一曲も追加されていないときは400",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().SetRepeatMode(gomock.Any(), false, "device_id").Return(nil)
				m.EXPECT().SetShuffleMode(gomock.Any(), false, "device_id").Return(nil)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:                     "sessionID",
					Name:                   "session_name",
					CreatorID:              "creator_id",
					QueueHead:              0,
					DeviceID:               "device_id",
					StateType:              entity.Stop,
					QueueTracks:            nil,
					AllowToControlByOthers: true,
				}, nil)

			},
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:      "StateType=STOP: 再生するデバイスがオフラインなら403",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().SetRepeatMode(gomock.Any(), false, "device_id").Return(entity.ErrActiveDeviceNotFound)
			},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Stop,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusForbidden,
		},
		{
			name:                  "指定されたidのセッションが存在しないとき404",
			sessionID:             "notFoundSessionID",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "notFoundSessionID").Return(nil, entity.ErrSessionNotFound)
			},
			wantErr:  true,
			wantCode: http.StatusNotFound,
		},
		{
			name:                  "StateType=ARCHIVED: 不正なstate遷移なので400",
			sessionID:             "sessionID",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{StateType: entity.Archived}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"state": "PLAY"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/sessions/:id/state")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)
			c = setToContext(c, tt.userID, nil)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			h := newSessionStateHandlerForTest(t, ctrl, tt.prepareMockPlayerFn, tt.prepareMockPusherFn,
				tt.prepareMockUserRepoFn, tt.prepareMockSessionRepoFn)

			err := h.State(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("State() error = %v, wantErr %v", err, tt.wantErr)
			}

			if er, ok := err.(*echo.HTTPError); ok && er.Code != tt.wantCode {
				t.Errorf("State() code = %d, want = %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestSessionHandler_State_PAUSE(t *testing.T) {
	tests := []struct {
		name                     string
		sessionID                string
		userID                   string
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		wantErr                  bool
		wantCode                 int
	}{
		{
			name:      "StateType=PLAY: 正しく一時停止処理が行われたら202",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Pause(gomock.Any(), "device_id").Return(nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{SessionID: "sessionID", Msg: entity.EventPause})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:                     "sessionID",
					Name:                   "session_name",
					CreatorID:              "creator_id",
					QueueHead:              0,
					DeviceID:               "device_id",
					StateType:              entity.Play,
					QueueTracks:            nil,
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:                     "sessionID",
					Name:                   "session_name",
					CreatorID:              "creator_id",
					QueueHead:              0,
					DeviceID:               "device_id",
					StateType:              entity.Pause,
					QueueTracks:            nil,
					AllowToControlByOthers: true,
				}).Return(nil)

			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:      "StateType=PLAY: 再生するデバイスがオフラインのときは、既に再生が止まっているはずなので202",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Pause(gomock.Any(), "device_id").Return(entity.ErrActiveDeviceNotFound)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{SessionID: "sessionID", Msg: entity.EventPause})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:                     "sessionID",
					Name:                   "session_name",
					CreatorID:              "creator_id",
					DeviceID:               "device_id",
					QueueHead:              0,
					StateType:              entity.Play,
					QueueTracks:            nil,
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:                     "sessionID",
					Name:                   "session_name",
					CreatorID:              "creator_id",
					DeviceID:               "device_id",
					QueueHead:              0,
					StateType:              entity.Pause,
					QueueTracks:            nil,
					AllowToControlByOthers: true,
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:      "StateType=PAUSE: 既にPAUSEでも一応一時停止APIを叩いて202",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Pause(gomock.Any(), "device_id").Return(nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{SessionID: "sessionID", Msg: entity.EventPause})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:                     "sessionID",
					Name:                   "session_name",
					CreatorID:              "creator_id",
					QueueHead:              0,
					DeviceID:               "device_id",
					StateType:              entity.Pause,
					QueueTracks:            nil,
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:                     "sessionID",
					Name:                   "session_name",
					CreatorID:              "creator_id",
					DeviceID:               "device_id",
					QueueHead:              0,
					StateType:              entity.Pause,
					QueueTracks:            nil,
					AllowToControlByOthers: true,
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:      "StateType=PAUSE: 既にPAUSEでも一応一時停止APIを叩くが、デバイスがオフラインなら既に再生が止まっているはずなので202",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Pause(gomock.Any(), "device_id").Return(entity.ErrActiveDeviceNotFound)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{SessionID: "sessionID", Msg: entity.EventPause})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:                     "sessionID",
					Name:                   "session_name",
					CreatorID:              "creator_id",
					QueueHead:              0,
					DeviceID:               "device_id",
					StateType:              entity.Pause,
					QueueTracks:            nil,
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:                     "sessionID",
					Name:                   "session_name",
					CreatorID:              "creator_id",
					DeviceID:               "device_id",
					QueueHead:              0,
					StateType:              entity.Pause,
					QueueTracks:            nil,
					AllowToControlByOthers: true,
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:                  "StateType=STOP: 不正なstate遷移なので400",
			sessionID:             "sessionID",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{StateType: entity.Stop}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:                  "StateType=ARCHIVED: 不正なstate遷移なので400",
			sessionID:             "sessionID",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{StateType: entity.Archived}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"state": "PAUSE"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/sessions/:id/state")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)
			c = setToContext(c, tt.userID, nil)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			h := newSessionStateHandlerForTest(t, ctrl, tt.prepareMockPlayerFn, tt.prepareMockPusherFn,
				tt.prepareMockUserRepoFn, tt.prepareMockSessionRepoFn)

			err := h.State(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("State() error = %v, wantErr %v", err, tt.wantErr)
			}

			if er, ok := err.(*echo.HTTPError); ok && er.Code != tt.wantCode {
				t.Errorf("State() code = %d, want = %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestSessionHandler_State_STOP(t *testing.T) {
	tests := []struct {
		name                     string
		sessionID                string
		userID                   string
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		wantErr                  bool
		wantCode                 int
	}{
		{
			name:                "StateType=ARCHIVED: 手動でアーカイブを解除して202",
			sessionID:           "sessionID",
			userID:              "creator_id",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{SessionID: "sessionID", Msg: entity.EventUnarchive})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: "ARCHIVED",
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().UpdateWithExpiredAt(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: "STOP",
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, gomock.Any()).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:                  "セッション作成者以外の操作のときは400",
			sessionID:             "sessionID",
			userID:                "non_creator_id",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{StateType: entity.Pause, CreatorID: "creator_id"}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:                  "StateType=PLAY: 不正なstateの変更なので400",
			sessionID:             "sessionID",
			userID:                "creator_id",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{StateType: entity.Play, CreatorID: "creator_id"}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:                  "StateType=PAUSE: 不正なstateの変更なので400",
			sessionID:             "sessionID",
			userID:                "creator_id",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{StateType: entity.Pause, CreatorID: "creator_id"}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
		{
			name:                  "StateType=STOP: なにもせずに202",
			sessionID:             "sessionID",
			userID:                "creator_id",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{StateType: entity.Stop, CreatorID: "creator_id"}, nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"state": "STOP"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/sessions/:id/state")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)
			c = setToContext(c, tt.userID, nil)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			h := newSessionStateHandlerForTest(t, ctrl, tt.prepareMockPlayerFn, tt.prepareMockPusherFn,
				tt.prepareMockUserRepoFn, tt.prepareMockSessionRepoFn)
			err := h.State(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("State() error = %v, wantErr %v", err, tt.wantErr)
			}

			if er, ok := err.(*echo.HTTPError); ok && er.Code != tt.wantCode {
				t.Errorf("State() code = %d, want = %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestSessionHandler_State_ARCHIVED(t *testing.T) {
	tests := []struct {
		name                     string
		sessionID                string
		userID                   string
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		wantErr                  bool
		wantCode                 int
	}{
		{
			name:      "StateType=PLAY: Spotifyでの再生を一時停止した後、正しくアーカイブされて202",
			sessionID: "sessionID",
			userID:    "creator_id",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Pause(gomock.Any(), "device_id").Return(nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{SessionID: "sessionID", Msg: entity.EventArchived})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Play,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Archived,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:                "StateType=PAUSE: 正しくアーカイブされて202",
			sessionID:           "sessionID",
			userID:              "creator_id",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{SessionID: "sessionID", Msg: entity.EventArchived})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Pause,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Archived,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:                "StateType=STOP: 正しくアーカイブされて202",
			sessionID:           "sessionID",
			userID:              "creator_id",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{SessionID: "sessionID", Msg: entity.EventArchived})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Stop,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: "ARCHIVED",
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:                  "StateType=ARCHIVED: 既にアーカイブされているので何もせずに202",
			sessionID:             "sessionID",
			userID:                "creator_id",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Archived,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, nil)
			},
			wantErr:  false,
			wantCode: http.StatusAccepted,
		},
		{
			name:                  "セッション作成者以外のときは400",
			sessionID:             "sessionID",
			userID:                "non_creator_id",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					DeviceID:  "device_id",
					StateType: entity.Archived,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
					AllowToControlByOthers: true,
				}, nil)
			},
			wantErr:  true,
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"state": "ARCHIVED"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/sessions/:id/state")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)
			c = setToContext(c, tt.userID, nil)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			h := newSessionStateHandlerForTest(t, ctrl, tt.prepareMockPlayerFn, tt.prepareMockPusherFn,
				tt.prepareMockUserRepoFn, tt.prepareMockSessionRepoFn)

			err := h.State(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("State() error = %v, wantErr %v", err, tt.wantErr)
			}

			if er, ok := err.(*echo.HTTPError); ok && er.Code != tt.wantCode {
				t.Errorf("State() code = %d, want = %d", rec.Code, tt.wantCode)
			}
		})
	}
}

// モックの準備
func newSessionStateHandlerForTest(
	t *testing.T,
	ctrl *gomock.Controller,
	prepareMockPlayerFn func(m *mock_spotify.MockPlayer),
	prepareMockPusherFn func(m *mock_event.MockPusher),
	prepareMockUserRepoFn func(m *mock_repository.MockUser),
	prepareMockSessionRepoFn func(m *mock_repository.MockSession)) *SessionHandler {
	t.Helper()

	mockPlayer := mock_spotify.NewMockPlayer(ctrl)
	prepareMockPlayerFn(mockPlayer)
	mockPusher := mock_event.NewMockPusher(ctrl)
	prepareMockPusherFn(mockPusher)
	mockUserRepo := mock_repository.NewMockUser(ctrl)
	prepareMockUserRepoFn(mockUserRepo)
	mockSessionRepo := mock_repository.NewMockSession(ctrl)
	prepareMockSessionRepoFn(mockSessionRepo)
	syncCheckTimerManager := entity.NewSyncCheckTimerManager()
	timerUC := usecase.NewSessionTimerUseCase(mockSessionRepo, mockPlayer, mockPusher, syncCheckTimerManager)
	uc := usecase.NewSessionUseCase(mockSessionRepo, mockUserRepo, mockPlayer, nil, nil, mockPusher, timerUC)
	stateUC := usecase.NewSessionStateUseCase(mockSessionRepo, mockPlayer, mockPusher, timerUC)
	return &SessionHandler{uc: uc, stateUC: stateUC}
}
