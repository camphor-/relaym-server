package usecase

import (
	"context"
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/mock_event"
	"github.com/camphor-/relaym-server/domain/mock_repository"
	"github.com/camphor-/relaym-server/domain/mock_spotify"

	"github.com/golang/mock/gomock"
)

func TestSessionUseCase_handleTrackEnd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		sessionID                string
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		wantTriggerAfterTrackEnd bool
		wantNextTrack            bool
		wantErr                  bool
	}{
		{
			name:                "最後の曲が再生し終わったときにSTOPイベントが送られる",
			sessionID:           "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.EventStop,
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Play,
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{},
						{},
					},
				}, nil)
				m.EXPECT().Update(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Stop,
					QueueHead: 2,
					QueueTracks: []*entity.QueueTrack{
						{},
						{},
					},
				}).Return(nil)
			},
			wantTriggerAfterTrackEnd: false,
			wantNextTrack:            false,
			wantErr:                  false,
		},
		{
			name:      "次の曲が存在するときはNEXTTRACKイベントが送られて、次の再生状態に遷移する",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().CurrentlyPlaying(gomock.Any()).Return(&entity.CurrentPlayingInfo{
					Playing:  true,
					Progress: 10000000,
					Track: &entity.Track{
						URI:      "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
						ID:       "06QTSGUEgcmKwiEJ0IMPig",
						Name:     "Borderland",
						Duration: 213066000000,
						Artists:  []*entity.Artist{{Name: "MONOEYES"}},
						URL:      "https://open.spotify.com/track/06QTSGUEgcmKwiEJ0IMPig",
						Album: &entity.Album{
							Name: "Interstate 46 E.P.",
							Images: []*entity.AlbumImage{
								{
									URL:    "https://i.scdn.co/image/ab67616d0000b273b48630d6efcebca2596120c4",
									Height: 640,
									Width:  640,
								},
							},
						},
					},
				}, nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.NewEventNextTrack(1),
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Play,
					QueueHead: 0,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:asfafefea",
							SessionID: "sessionID",
						},
						{
							Index:     1,
							URI:       "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
							SessionID: "sessionID",
						},
					},
				}, nil)
				m.EXPECT().Update(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Play,
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:asfafefea",
							SessionID: "sessionID",
						},
						{
							Index:     1,
							URI:       "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
							SessionID: "sessionID",
						},
					},
				}).Return(nil)
			},
			wantTriggerAfterTrackEnd: true,
			wantNextTrack:            true,
			wantErr:                  false,
		},
		{
			name:      "次の曲が存在し、次に再生される曲の二曲先の曲が存在するときはNEXTTRACKイベントが送られて、次の再生状態に遷移し、同時に二曲先の曲がSpotifyのqueueに積まれる",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().CurrentlyPlaying(gomock.Any()).Return(&entity.CurrentPlayingInfo{
					Playing:  true,
					Progress: 10000000,
					Track: &entity.Track{
						URI:      "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
						ID:       "06QTSGUEgcmKwiEJ0IMPig",
						Name:     "Borderland",
						Duration: 213066000000,
						Artists:  []*entity.Artist{{Name: "MONOEYES"}},
						URL:      "https://open.spotify.com/track/06QTSGUEgcmKwiEJ0IMPig",
						Album: &entity.Album{
							Name: "Interstate 46 E.P.",
							Images: []*entity.AlbumImage{
								{
									URL:    "https://i.scdn.co/image/ab67616d0000b273b48630d6efcebca2596120c4",
									Height: 640,
									Width:  640,
								},
							},
						},
					},
				}, nil)
				m.EXPECT().Enqueue(gomock.Any(), "spotify:track:3", "deviceID").Return(nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.NewEventNextTrack(1),
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Play,
					QueueHead: 0,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:asfafefea",
							SessionID: "sessionID",
						},
						{
							Index:     1,
							URI:       "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
							SessionID: "sessionID",
						},
						{
							Index:     2,
							URI:       "spotify:track:2",
							SessionID: "sessionID",
						},
						{
							Index:     3,
							URI:       "spotify:track:3",
							SessionID: "sessionID",
						},
					},
				}, nil)
				m.EXPECT().Update(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Play,
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:asfafefea",
							SessionID: "sessionID",
						},
						{
							Index:     1,
							URI:       "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
							SessionID: "sessionID",
						},
						{
							Index:     2,
							URI:       "spotify:track:2",
							SessionID: "sessionID",
						},
						{
							Index:     3,
							URI:       "spotify:track:3",
							SessionID: "sessionID",
						},
					},
				}).Return(nil)
			},
			wantTriggerAfterTrackEnd: true,
			wantNextTrack:            true,
			wantErr:                  false,
		},
		{
			name:      "次の曲が存在するが、実際には違う曲が流れていた場合はINTERRUPTイベントが送られる",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().CurrentlyPlaying(gomock.Any()).Return(&entity.CurrentPlayingInfo{
					Playing:  true,
					Progress: 10000000,
					Track: &entity.Track{
						URI:      "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
						ID:       "06QTSGUEgcmKwiEJ0IMPig",
						Name:     "Borderland",
						Duration: 213066000000,
						Artists:  []*entity.Artist{{Name: "MONOEYES"}},
						URL:      "https://open.spotify.com/track/06QTSGUEgcmKwiEJ0IMPig",
						Album: &entity.Album{
							Name: "Interstate 46 E.P.",
							Images: []*entity.AlbumImage{
								{
									URL:    "https://i.scdn.co/image/ab67616d0000b273b48630d6efcebca2596120c4",
									Height: 640,
									Width:  640,
								},
							},
						},
					},
				}, nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.EventInterrupt,
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Play,
					QueueHead: 0,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:asfafefea",
							SessionID: "sessionID",
						},
						{
							Index:     1,
							URI:       "spotify:track:differentTrack",
							SessionID: "sessionID",
						},
					},
				}, nil)
				m.EXPECT().Update(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Stop,
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:asfafefea",
							SessionID: "sessionID",
						},
						{
							Index:     1,
							URI:       "spotify:track:differentTrack",
							SessionID: "sessionID",
						},
					},
				}).Return(nil)
			},
			wantTriggerAfterTrackEnd: false,
			wantNextTrack:            false,
			wantErr:                  true,
		},
		{
			name:      "次の曲が存在するが、デバイスがオフラインになっていた場合はINTERRUPTイベントが送られる",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
				m.EXPECT().CurrentlyPlaying(gomock.Any()).Return(nil, entity.ErrActiveDeviceNotFound)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.EventInterrupt,
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Play,
					QueueHead: 0,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:asfafefea",
							SessionID: "sessionID",
						},
						{
							Index:     1,
							URI:       "spotify:track:differentTrack",
							SessionID: "sessionID",
						},
					},
				}, nil)
				m.EXPECT().Update(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Stop,
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:asfafefea",
							SessionID: "sessionID",
						},
						{
							Index:     1,
							URI:       "spotify:track:differentTrack",
							SessionID: "sessionID",
						},
					},
				}).Return(nil)
			},
			wantTriggerAfterTrackEnd: false,
			wantNextTrack:            false,
			wantErr:                  true,
		},
		{
			name:      "呼び出された時点でsessionのstateがARCHIVEDになっていた時にはtimerをdeleteしてArchivedのイベントを送信する",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.EventArchived,
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Archived,
					QueueHead: 0,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:asfafefea",
							SessionID: "sessionID",
						},
						{
							Index:     1,
							URI:       "spotify:track:differentTrack",
							SessionID: "sessionID",
						},
					},
				}, nil)
				m.EXPECT().Update(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: entity.Archived,
					QueueHead: 0,
					QueueTracks: []*entity.QueueTrack{
						{
							Index:     0,
							URI:       "spotify:track:asfafefea",
							SessionID: "sessionID",
						},
						{
							Index:     1,
							URI:       "spotify:track:differentTrack",
							SessionID: "sessionID",
						},
					},
				}).Return(nil)
			},
			wantTriggerAfterTrackEnd: false,
			wantNextTrack:            false,
			wantErr:                  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			s := NewSessionUseCase(mockSessionRepo, mockUserRepo, mockPlayer, nil, nil, mockPusher)

			gotTriggerAfterTrackEnd, gotNextTrack, err := s.handleTrackEnd(context.Background(), tt.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleTrackEnd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotTriggerAfterTrackEnd != nil) != tt.wantTriggerAfterTrackEnd {
				t.Errorf("handleTrackEnd() gotTriggerAfterTrackEnd = %v, want %v", gotTriggerAfterTrackEnd, tt.wantTriggerAfterTrackEnd)
			}
			if gotNextTrack != tt.wantNextTrack {
				t.Errorf("handleTrackEnd() gotNextTrack = %v, want %v", gotNextTrack, tt.wantNextTrack)
			}
		})
	}
}

func TestSessionUseCase_CanConnectToPusher(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		sessionID                string
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		wantErr                  bool
	}{
		{
			name:      "存在しないセッションのとき404",
			sessionID: "not_found_session_id",
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("not_found_session_id").Return(nil, entity.ErrSessionNotFound)
			},
			wantErr: true,
		},
		{
			name:      "StateがStopのセッションのとき正しくWebSocketのコネクションが確立される",
			sessionID: "sessionID",
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:          "sessionID",
					Name:        "session_name",
					CreatorID:   "creator_id",
					QueueHead:   0,
					StateType:   "STOP",
					QueueTracks: []*entity.QueueTrack{},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:      "StateがPlayのセッションでタイマーが存在しないので、タイマーを作成した後、正しくWebSocketのコネクションが確立される",
			sessionID: "sessionID",
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "session_name",
					CreatorID: "creator_id",
					QueueHead: 0,
					StateType: "PLAY",
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:49BRCNV7E94s7Q2FUhhT3w"},
					},
				}, nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockSessionRepo := mock_repository.NewMockSession(ctrl)
			tt.prepareMockSessionRepoFn(mockSessionRepo)
			s := NewSessionUseCase(mockSessionRepo, nil, nil, nil, nil, nil)

			if err := s.CanConnectToPusher(context.Background(), tt.sessionID); (err != nil) != tt.wantErr {
				t.Errorf("CanConnectToPusher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
