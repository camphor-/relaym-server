package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
				m.EXPECT().FindByID(gomock.Any(), "session_id").Return(nil, entity.ErrSessionNotFound)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			wantErr:               true,
			wantCode:              http.StatusNotFound,
		},
		{
			name:      "正しくデバイスをセットできると204",
			userID:    "creator_id",
			sessionID: "session_id",
			body:      `{"device_id": "device_id"}`,
			prepareMockRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "session_id").Return(&entity.Session{
					ID:          "session_id",
					Name:        "name",
					CreatorID:   "creator_id",
					DeviceID:    "",
					StateType:   "PAUSE",
					QueueHead:   0,
					QueueTracks: nil,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
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

			uc := usecase.NewSessionUseCase(mockRepo, mockUserRepo, nil, nil, nil, nil, nil)
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
		ID:                     "ID",
		Name:                   "go! go! session!",
		AllowToControlByOthers: true,
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
			Tracks: []*trackJSON{},
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
			body:                `{"name": "go! go! session!", "allow_to_control_by_others": true}`,
			userID:              "creatorID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().StoreSession(gomock.Any(), gomock.Any()).Return(nil)
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
			body:                     `{"name": "", "allow_to_control_by_others": true}`,
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
			h := newSessionStateHandlerForTest(t, ctrl, tt.prepareMockPlayerFn, tt.prepareMockPusherFn, tt.prepareMockUserRepoFn, tt.prepareMockSessionRepoFn)

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

func TestSessionHandler_Enqueue(t *testing.T) {
	session := &entity.Session{
		ID:        "sessionID",
		Name:      "sessionName",
		CreatorID: "sessionCreator",
		DeviceID:  "sessionDeviceID",
		StateType: "PLAY",
		QueueHead: 0,
		QueueTracks: []*entity.QueueTrack{{
			Index:     0,
			URI:       "spotify:track:track_uri",
			SessionID: "sessionID",
		},
		},
	}
	sessionHadManyTracks := &entity.Session{
		ID:        "sessionHadManyTracksID",
		Name:      "sessionName",
		CreatorID: "sessionCreator",
		DeviceID:  "sessionDeviceID",
		StateType: "PLAY",
		QueueHead: 0,
		QueueTracks: []*entity.QueueTrack{
			{
				Index:     0,
				URI:       "spotify:track:track_uri1",
				SessionID: "sessionID",
			},
			{
				Index:     1,
				URI:       "spotify:track:track_uri2",
				SessionID: "sessionID",
			},
			{
				Index:     2,
				URI:       "spotify:track:track_uri3",
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
			name:                "正しいuriが渡されると正常に動作し、sessionがqueueの最後から二番目以内ではない曲を再生している場合はEnqueueを叩かない",
			sessionID:           "sessionHadManyTracksID",
			body:                `{"uri": "spotify:track:valid_uri"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionHadManyTracksID",
					Msg:       entity.EventAddTrack,
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionHadManyTracksID").Return(sessionHadManyTracks, nil)
				m.EXPECT().StoreQueueTrack(gomock.Any(), &entity.QueueTrackToStore{
					URI:       "spotify:track:valid_uri",
					SessionID: "sessionHadManyTracksID",
				}).Return(nil)
			},
			wantErr:  false,
			wantCode: http.StatusNoContent,
		},
		{
			name:      "正しくuriが渡されると正常に動作し、sessionがqueueの最後から二番目以内の曲を再生している場合はEnqueueを叩く",
			sessionID: "sessionID",
			body:      `{"uri": "spotify:track:valid_uri"}`,
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().Enqueue(gomock.Any(), "spotify:track:valid_uri", "sessionDeviceID").Return(nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.EventAddTrack,
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(session, nil)
				m.EXPECT().StoreQueueTrack(gomock.Any(), &entity.QueueTrackToStore{
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
				m.EXPECT().FindByID(gomock.Any(), "invalidSessionID").Return(nil, entity.ErrSessionNotFound)
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
			h := newSessionStateHandlerForTest(t, ctrl, tt.prepareMockPlayerFn, tt.prepareMockPusherFn, tt.prepareMockUserRepoFn, tt.prepareMockSessionRepoFn)

			err := h.Enqueue(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Enqueue() error = %v, wantErr %v", err, tt.wantErr)
			}

			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); ok && er.Code != tt.wantCode {
				t.Errorf("Enqueue() code = %d, want = %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestSessionHandler_GetSession(t *testing.T) {
	session := &entity.Session{
		ID:        "sessionID",
		Name:      "sessionName",
		CreatorID: "creatorID",
		DeviceID:  "sessionDeviceID",
		StateType: "STOP",
		QueueHead: 0,
		QueueTracks: []*entity.QueueTrack{
			{
				Index:     0,
				URI:       "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
				SessionID: "sessionID",
			},
		},
	}
	user := &entity.User{
		ID:            "creatorID",
		SpotifyUserID: "creatorSpotifyUserID",
		DisplayName:   "creatorDisplayName",
	}
	artists := []*entity.Artist{
		{
			Name: "MONOEYES",
		},
	}

	albumImages := []*entity.AlbumImage{
		{
			URL:    "https://i.scdn.co/image/ab67616d0000b273b48630d6efcebca2596120c4",
			Height: 640,
			Width:  640,
		},
	}

	tracks := []*entity.Track{
		{
			URI:      "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
			ID:       "06QTSGUEgcmKwiEJ0IMPig",
			Name:     "Borderland",
			Duration: 213066000000,
			Artists:  artists,
			URL:      "https://open.spotify.com/track/06QTSGUEgcmKwiEJ0IMPig",
			Album: &entity.Album{
				Name:   "Interstate 46 E.P.",
				Images: albumImages,
			},
		},
	}
	device := &entity.Device{
		ID:           "sessionDeviceID",
		IsRestricted: false,
		Name:         "hogeさんのiPhone11",
	}

	cpi := &entity.CurrentPlayingInfo{
		Playing:  false,
		Progress: 0,
		Track:    tracks[0],
		Device:   device,
	}

	albumImageJSONs := []*albumImageJSON{
		{
			URL:    "https://i.scdn.co/image/ab67616d0000b273b48630d6efcebca2596120c4",
			Height: 640,
			Width:  640,
		},
	}

	artistJSONs := []*artistJSON{
		{
			Name: "MONOEYES",
		},
	}

	trackJSONs := []*trackJSON{
		{
			URI:      "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
			ID:       "06QTSGUEgcmKwiEJ0IMPig",
			Name:     "Borderland",
			Duration: 213066,
			Artists:  artistJSONs,
			URL:      "https://open.spotify.com/track/06QTSGUEgcmKwiEJ0IMPig",
			Album: &albumJSON{
				Name:   "Interstate 46 E.P.",
				Images: albumImageJSONs,
			},
		},
	}

	sessionResponse := &sessionRes{
		ID:   "sessionID",
		Name: "sessionName",
		Creator: creatorJSON{
			ID:          "creatorID",
			DisplayName: "creatorDisplayName",
		},
		Playback: playbackJSON{
			State: stateJSON{
				Type: "STOP",
			},
			Device: &deviceJSON{
				ID:           device.ID,
				IsRestricted: device.IsRestricted,
				Name:         device.Name,
			},
		},
		Queue: queueJSON{
			Head:   0,
			Tracks: trackJSONs,
		},
	}

	tests := []struct {
		name                     string
		sessionID                string
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockTrackCliFn    func(m *mock_spotify.MockTrackClient)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		addToTimerSessionID      string
		want                     *sessionRes
		wantErr                  bool
		wantCode                 int
	}{
		{
			name:      "与えられたIDのsessionが存在するとき正常に動作する",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().CurrentlyPlaying(gomock.Any()).Return(cpi, nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			prepareMockTrackCliFn: func(m *mock_spotify.MockTrackClient) {
				m.EXPECT().GetTracksFromURI(gomock.Any(), []string{"spotify:track:06QTSGUEgcmKwiEJ0IMPig"}).Return(tracks, nil)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {
				m.EXPECT().FindByID("creatorID").Return(user, nil)
			},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "sessionID").Return(session, nil)
			},
			want:     sessionResponse,
			wantErr:  false,
			wantCode: http.StatusOK,
		},
		{
			name:      "存在しないsessionIDを渡した時404",
			sessionID: "non_exist_sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			prepareMockTrackCliFn: func(m *mock_spotify.MockTrackClient) {
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {
			},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "non_exist_sessionID").Return(nil, entity.ErrSessionNotFound)
			},
			want:     nil,
			wantErr:  true,
			wantCode: http.StatusNotFound,
		},
		{
			name:      "セッションがStop以外のときはStateに再生状況が含まれる",
			sessionID: "play_sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().CurrentlyPlaying(gomock.Any()).Return(cpi, nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {},
			prepareMockTrackCliFn: func(m *mock_spotify.MockTrackClient) {
				m.EXPECT().GetTracksFromURI(gomock.Any(), []string{"spotify:track:06QTSGUEgcmKwiEJ0IMPig"}).Return(tracks, nil)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {
				m.EXPECT().FindByID("creatorID").Return(user, nil)
			},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "play_sessionID").Return(&entity.Session{
					ID:        "play_sessionID",
					Name:      "sessionName",
					CreatorID: "creatorID",
					DeviceID:  "sessionDeviceID",
					StateType: "PLAY",
					QueueHead: 0,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
							SessionID: "sessionID",
						},
					},
				}, nil)
			},
			want: &sessionRes{
				ID:   "play_sessionID",
				Name: "sessionName",
				Creator: creatorJSON{
					ID:          "creatorID",
					DisplayName: "creatorDisplayName",
				},
				Playback: playbackJSON{
					State: stateJSON{
						Type:      "PLAY",
						Length:    convToPointer(tracks[0].Duration.Milliseconds()),
						Progress:  convToPointer(0),
						Remaining: convToPointer(tracks[0].Duration.Milliseconds()),
					},
					Device: &deviceJSON{
						ID:           device.ID,
						IsRestricted: device.IsRestricted,
						Name:         device.Name,
					},
				},
				Queue: queueJSON{
					Head:   0,
					Tracks: trackJSONs,
				},
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
		{
			name:      "セッションがStop以外のときは同期チェックに失敗するとStopになる",
			sessionID: "play_sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().CurrentlyPlaying(gomock.Any()).Return(&entity.CurrentPlayingInfo{
					Playing:  false,
					Progress: 0,
					Track: &entity.Track{
						URI:      "spotify:track:another_track",
						ID:       "another_track",
						Name:     "勝手にSpotifyで再生された曲",
						Duration: 213066000000,
						Artists:  artists,
						URL:      "https://open.spotify.com/track/06QTSGUEgcmKwiEJ0IMPig",
						Album: &entity.Album{
							Name:   "Interstate 46 E.P.",
							Images: albumImages,
						},
					},
					Device: device,
				}, nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "play_sessionID",
					Msg:       entity.EventInterrupt,
				})
			},
			prepareMockTrackCliFn: func(m *mock_spotify.MockTrackClient) {
				m.EXPECT().GetTracksFromURI(gomock.Any(), []string{"spotify:track:06QTSGUEgcmKwiEJ0IMPig"}).Return(tracks, nil)
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {
				m.EXPECT().FindByID("creatorID").Return(user, nil)
			},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID(gomock.Any(), "play_sessionID").Return(&entity.Session{
					ID:        "play_sessionID",
					Name:      "sessionName",
					CreatorID: "creatorID",
					DeviceID:  "sessionDeviceID",
					StateType: "PLAY",
					QueueHead: 0,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
							SessionID: "sessionID",
						},
					},
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "play_sessionID",
					Name:      "sessionName",
					CreatorID: "creatorID",
					DeviceID:  "sessionDeviceID",
					StateType: "STOP",
					QueueHead: 0,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
							SessionID: "sessionID",
						},
					},
				}).Return(nil)
			},
			addToTimerSessionID: "play_sessionID",
			want: &sessionRes{
				ID:   "play_sessionID",
				Name: "sessionName",
				Creator: creatorJSON{
					ID:          "creatorID",
					DisplayName: "creatorDisplayName",
				},
				Playback: playbackJSON{
					State: stateJSON{
						Type: "STOP",
					},
					Device: &deviceJSON{
						ID:           device.ID,
						IsRestricted: device.IsRestricted,
						Name:         device.Name,
					},
				},
				Queue: queueJSON{
					Head:   0,
					Tracks: trackJSONs,
				},
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/sessions/:id")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			h := newSessionHandlerForTest(t, ctrl, tt.prepareMockPlayerFn, tt.prepareMockTrackCliFn, tt.prepareMockPusherFn, tt.prepareMockUserRepoFn, tt.prepareMockSessionRepoFn, tt.addToTimerSessionID)
			err := h.GetSession(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("GetSession() code = %d, want = %d", rec.Code, tt.wantCode)
				return
			}

			if !tt.wantErr {
				got := &sessionRes{}
				err := json.Unmarshal(rec.Body.Bytes(), got)
				if err != nil {
					t.Fatal(err)
					return
				}
				opts := []cmp.Option{cmpopts.IgnoreFields(sessionRes{}, "ID")}
				if !cmp.Equal(got, tt.want, opts...) {
					t.Errorf("GetSession() diff = %v", cmp.Diff(got, tt.want, opts...))
				}
			}
		})
	}
}

func TestUserHandler_GetActiveDevices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		prepareMockUserSpo func(mock *mock_spotify.MockUser)
		want               *devicesRes
		wantErr            bool
		wantCode           int
	}{
		{
			name: "正しくデバイスを取得できる",
			prepareMockUserSpo: func(mock *mock_spotify.MockUser) {
				devices := []*entity.Device{
					{
						ID:           "hoge_id",
						IsRestricted: false,
						Name:         "hogeさんのiPhone11",
					},
				}
				mock.EXPECT().GetActiveDevices(gomock.Any()).Return(devices, nil)
			},
			want: &devicesRes{
				Devices: []*deviceJSON{
					{
						ID:           "hoge_id",
						IsRestricted: false,
						Name:         "hogeさんのiPhone11",
					},
				},
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
		{
			name: "spotify.GetActiveDevicesが失敗した時にInternalServerErrorが返る",
			prepareMockUserSpo: func(mock *mock_spotify.MockUser) {
				mock.EXPECT().GetActiveDevices(gomock.Any()).Return(nil, errors.New("unknown error"))
			},
			want:     nil,
			wantErr:  true,
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mock := mock_spotify.NewMockUser(ctrl)
			tt.prepareMockUserSpo(mock)
			uc := usecase.NewSessionUseCase(nil, nil, nil, nil, mock, nil, nil)
			h := &SessionHandler{uc: uc}

			err := h.GetActiveDevices(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActiveDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("GetActiveDevices() code = %d, want = %d", rec.Code, tt.wantCode)
			}

			if !tt.wantErr {
				got := &devicesRes{}
				err := json.Unmarshal(rec.Body.Bytes(), got)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, tt.want) {
					t.Errorf("GetActiveDevices() diff = %v", cmp.Diff(got, tt.want))
				}
			}
		})
	}
}

func TestUserHandler_NextTrack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		sessionID                string
		userID                   string
		prepareMockPlayerCliFn   func(m *mock_spotify.MockPlayer)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockTrackCli      func(m *mock_spotify.MockTrackClient)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		wantErr                  bool
		wantCode                 int
	}{
		{
			name:      "Playかつ次の曲が存在する時に次の曲にPlayのまま遷移, 202",
			sessionID: "sessionID",
			userID:    "userID",
			wantErr:   false,
			wantCode:  http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/sessions/:id/devices")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)
			c = setToContext(c, tt.userID, nil)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockPC := mock_spotify.NewMockPlayer(ctrl)
			tt.prepareMockPlayerCliFn(mockPC)
			mockP := mock_event.NewMockPusher(ctrl)
			tt.prepareMockPusherFn(mockP)
			mockSR := mock_repository.NewMockSession(ctrl)
			tt.prepareMockSessionRepoFn(mockSR)
			mockTC := mock_spotify.NewMockTrackClient(ctrl)
			tt.prepareMockTrackCli(mockTC)
			mockUR := mock_repository.NewMockUser(ctrl)
			tt.prepareMockUserRepoFn(mockUR)
			syncCheckTimerManager := entity.NewSyncCheckTimerManager()
			icm := entity.NewInterruptChanManager()
			timerUC := usecase.NewSessionTimerUseCase(mockSR, mockPC, mockP, syncCheckTimerManager, icm)

			uc := usecase.NewSessionUseCase(mockSR, mockUR, mockPC, mockTC, nil, mockP, timerUC)
			h := &SessionHandler{uc: uc}

			err := h.GetActiveDevices(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("NextTrack() error = %v, wantErr %v", err, tt.wantErr)
			}
			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("NextTrack() code = %d, want = %d", rec.Code, tt.wantCode)
			}

		})
	}
}

func convToPointer(given int64) *int64 {
	return &given
}

// モックの準備
func newSessionHandlerForTest(
	t *testing.T,
	ctrl *gomock.Controller,
	prepareMockPlayerFn func(m *mock_spotify.MockPlayer),
	prepareMockTrackFun func(m *mock_spotify.MockTrackClient),
	prepareMockPusherFn func(m *mock_event.MockPusher),
	prepareMockUserRepoFn func(m *mock_repository.MockUser),
	prepareMockSessionRepoFn func(m *mock_repository.MockSession),
	sessionID string) *SessionHandler {
	t.Helper()

	mockPlayer := mock_spotify.NewMockPlayer(ctrl)
	prepareMockPlayerFn(mockPlayer)
	mockTrackCli := mock_spotify.NewMockTrackClient(ctrl)
	prepareMockTrackFun(mockTrackCli)
	mockPusher := mock_event.NewMockPusher(ctrl)
	prepareMockPusherFn(mockPusher)
	mockUserRepo := mock_repository.NewMockUser(ctrl)
	prepareMockUserRepoFn(mockUserRepo)
	mockSessionRepo := mock_repository.NewMockSession(ctrl)
	prepareMockSessionRepoFn(mockSessionRepo)
	syncCheckTimerManager := entity.NewSyncCheckTimerManager()
	icm := entity.NewInterruptChanManager()
	if sessionID != "" {
		timer := syncCheckTimerManager.CreateTimer(sessionID)
		timer.SetTimer(5 * time.Minute)
	}
	timerUC := usecase.NewSessionTimerUseCase(mockSessionRepo, mockPlayer, mockPusher, syncCheckTimerManager, icm)
	uc := usecase.NewSessionUseCase(mockSessionRepo, mockUserRepo, mockPlayer, mockTrackCli, nil, mockPusher, timerUC)
	return &SessionHandler{uc: uc}
}
