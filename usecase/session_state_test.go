package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/camphor-/relaym-server/domain/service"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/mock_event"
	"github.com/camphor-/relaym-server/domain/mock_repository"
	"github.com/camphor-/relaym-server/domain/mock_spotify"
	"github.com/golang/mock/gomock"
)

func TestSessionStateUseCase_nextTrackInPauseTx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		sessionID                string
		userID                   string
		addToTimerSessionID      string
		prepareMockPlayerCliFn   func(m *mock_spotify.MockPlayer)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockTrackCliFn    func(m *mock_spotify.MockTrackClient)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		wantErr                  bool
	}{
		{
			name:                "Pauseかつ次の曲が存在すると次の曲に遷移し、202",
			sessionID:           "sessionID",
			userID:              "userID",
			addToTimerSessionID: "sessionID",
			prepareMockPlayerCliFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().GoNextTrack(gomock.Any(), "deviceID").Return(nil)
				m.EXPECT().Pause(gomock.Any(), "deviceID").Return(nil)
			},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(
					&entity.Session{
						ID:        "sessionID",
						Name:      "name",
						CreatorID: "creatorID",
						DeviceID:  "deviceID",
						StateType: "PAUSE",
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
						},
						ExpiredAt:              time.Time{},
						AllowToControlByOthers: true,
						ProgressWhenPaused:     0,
					}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: "PAUSE",
					QueueHead: 1,
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
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: true,
					ProgressWhenPaused:     0,
				})
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.NewEventNextTrack(1),
				})
			},
			prepareMockTrackCliFn: func(m *mock_spotify.MockTrackClient) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			wantErr:               false,
		},
		{
			name:                "Pauseかつ次の曲が3曲存在すると次の曲に遷移し、三曲先がEnqueueされ、202",
			sessionID:           "sessionID",
			userID:              "userID",
			addToTimerSessionID: "sessionID",
			prepareMockPlayerCliFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().GoNextTrack(gomock.Any(), "deviceID").Return(nil)
				m.EXPECT().Pause(gomock.Any(), "deviceID").Return(nil)
				m.EXPECT().Enqueue(gomock.Any(), "spotify:track:track_uri4", "deviceID").Return(nil)
			},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(
					&entity.Session{
						ID:        "sessionID",
						Name:      "name",
						CreatorID: "creatorID",
						DeviceID:  "deviceID",
						StateType: "PAUSE",
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
							{
								Index:     3,
								URI:       "spotify:track:track_uri4",
								SessionID: "sessionID",
							},
						},
						ExpiredAt:              time.Time{},
						AllowToControlByOthers: true,
						ProgressWhenPaused:     0,
					}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: "PAUSE",
					QueueHead: 1,
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
						{
							Index:     3,
							URI:       "spotify:track:track_uri4",
							SessionID: "sessionID",
						},
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: true,
					ProgressWhenPaused:     0,
				})
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.NewEventNextTrack(1),
				})
			},
			prepareMockTrackCliFn: func(m *mock_spotify.MockTrackClient) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			wantErr:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := newSessionStateUseCaseForTest(t, ctrl, tt.prepareMockPlayerCliFn, tt.prepareMockTrackCliFn,
				tt.prepareMockPusherFn, tt.prepareMockUserRepoFn, tt.prepareMockSessionRepoFn, tt.addToTimerSessionID)

			ctx := context.Background()
			ctx = service.SetUserIDToContext(ctx, tt.userID)

			_, err := uc.nextTrackInPauseTx(tt.sessionID)(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NextTrack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSessionStateUseCase_nextTrackInStopTx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		sessionID                string
		userID                   string
		addToTimerSessionID      string
		prepareMockPlayerCliFn   func(m *mock_spotify.MockPlayer)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockTrackCliFn    func(m *mock_spotify.MockTrackClient)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		wantErr                  bool
	}{
		{
			name:                   "STOPかつ次の曲が存在する時に次の曲にSTOPのまま遷移,202",
			sessionID:              "sessionID",
			userID:                 "userID",
			addToTimerSessionID:    "sessionID",
			prepareMockPlayerCliFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(
					&entity.Session{
						ID:        "sessionID",
						Name:      "name",
						CreatorID: "creatorID",
						DeviceID:  "deviceID",
						StateType: "STOP",
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
						},
						ExpiredAt:              time.Time{},
						AllowToControlByOthers: true,
						ProgressWhenPaused:     0,
					}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: "STOP",
					QueueHead: 1,
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
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: true,
					ProgressWhenPaused:     0,
				}).Return(nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.NewEventNextTrack(1),
				})
			},
			prepareMockTrackCliFn: func(m *mock_spotify.MockTrackClient) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			wantErr:               false,
		},
		{
			name:                   "STOPかつ次の曲が存在しない時にErrNextQueueTrackNotFound,400",
			sessionID:              "sessionID",
			userID:                 "userID",
			addToTimerSessionID:    "sessionID",
			prepareMockPlayerCliFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(
					&entity.Session{
						ID:        "sessionID",
						Name:      "name",
						CreatorID: "creatorID",
						DeviceID:  "deviceID",
						StateType: "STOP",
						QueueHead: 2,
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
						},
						ExpiredAt:              time.Time{},
						AllowToControlByOthers: true,
						ProgressWhenPaused:     0,
					}, nil)
			},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockTrackCliFn: func(m *mock_spotify.MockTrackClient) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			wantErr:               true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := newSessionStateUseCaseForTest(t, ctrl, tt.prepareMockPlayerCliFn, tt.prepareMockTrackCliFn,
				tt.prepareMockPusherFn, tt.prepareMockUserRepoFn, tt.prepareMockSessionRepoFn, tt.addToTimerSessionID)

			ctx := context.Background()
			ctx = service.SetUserIDToContext(ctx, tt.userID)

			_, err := uc.nextTrackInStopTx(tt.sessionID)(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NextTrack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// モックの準備
func newSessionStateUseCaseForTest(
	t *testing.T,
	ctrl *gomock.Controller,
	prepareMockPlayerFn func(m *mock_spotify.MockPlayer),
	prepareMockTrackFun func(m *mock_spotify.MockTrackClient),
	prepareMockPusherFn func(m *mock_event.MockPusher),
	prepareMockUserRepoFn func(m *mock_repository.MockUser),
	prepareMockSessionRepoFn func(m *mock_repository.MockSession),
	sessionID string) *SessionStateUseCase {
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
	if sessionID != "" {
		timer := syncCheckTimerManager.CreateExpiredTimer(sessionID)
		timer.SetDuration(5 * time.Minute)
	}
	timerUC := NewSessionTimerUseCase(mockSessionRepo, mockPlayer, mockPusher, syncCheckTimerManager)
	return NewSessionStateUseCase(mockSessionRepo, mockPlayer, mockPusher, timerUC)

}
