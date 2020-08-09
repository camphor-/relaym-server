package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/mock_event"
	"github.com/camphor-/relaym-server/domain/mock_repository"
	"github.com/camphor-/relaym-server/domain/mock_spotify"
	"github.com/golang/mock/gomock"
)

func TestSessionTimerUseCase_handleTrackEndTx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		sessionID                string
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
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
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(&entity.Session{
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
				m.EXPECT().Update(gomock.Any(), &entity.Session{
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
			wantNextTrack: false,
			wantErr:       false,
		},
		{
			name:      "次の曲が存在するときはNEXTTRACKイベントが送られて、次の再生状態に遷移する",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(&entity.Session{
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
				m.EXPECT().Update(gomock.Any(), &entity.Session{
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
			wantNextTrack: true,
			wantErr:       false,
		},
		{
			name:                  "次の曲が存在するが、実際には違う曲が流れていた場合はINTERRUPTイベントが送られる",
			sessionID:             "sessionID",
			prepareMockPlayerFn:   func(m *mock_spotify.MockPlayer) {},
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(&entity.Session{
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
				m.EXPECT().Update(gomock.Any(), &entity.Session{
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
							URI:       "spotify:track:differentTrack",
							SessionID: "sessionID",
						},
					},
				}).Return(nil)
			},
			wantNextTrack: true,
			wantErr:       false,
		},
		{
			name:      "次の曲が存在するが、デバイスがオフラインになっていた場合はINTERRUPTイベントが送られる",
			sessionID: "sessionID",
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(&entity.Session{
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
				m.EXPECT().Update(gomock.Any(), &entity.Session{
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
							URI:       "spotify:track:differentTrack",
							SessionID: "sessionID",
						},
					},
				}).Return(nil)
			},
			wantNextTrack: true,
			wantErr:       false,
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
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(&entity.Session{
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
				m.EXPECT().Update(gomock.Any(), &entity.Session{
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
			wantNextTrack: false,
			wantErr:       false,
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

			tmpWaitTimeBeforeHandleTrackEnd := waitTimeAfterHandleTrackEnd
			waitTimeAfterHandleTrackEnd = 0
			defer func() {
				waitTimeAfterHandleTrackEnd = tmpWaitTimeBeforeHandleTrackEnd
			}()

			syncCheckTimerManager := entity.NewSyncCheckTimerManager()

			s := NewSessionTimerUseCase(mockSessionRepo, mockPlayer, mockPusher, syncCheckTimerManager)
			gotTriggerAfterTrackEndResponseInterface, err := s.handleTrackEndTx(tt.sessionID)(context.Background())

			gotHandleTrackEndResponse, ok := gotTriggerAfterTrackEndResponseInterface.(*handleTrackEndResponse)
			if !ok {
				t.Fatal("gotTriggerAfterTrackEndResponse should be *handleTrackEndResponse")
			}
			if (gotHandleTrackEndResponse.err != nil) != tt.wantErr {
				t.Errorf("handleTrackEnd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHandleTrackEndResponse.nextTrack != tt.wantNextTrack {
				t.Errorf("handleTrackEnd() gotNextTrack = %v, want %v", gotHandleTrackEndResponse.nextTrack, tt.wantNextTrack)
			}
		})
	}
}

func TestSessionTimerUseCase_handleWaitTimerExpiredTx(t *testing.T) {
	tests := []struct {
		name                     string
		sessionID                string
		currentOperation         currentOperation
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		prepareMockPusherFn      func(m *mock_event.MockPusher)
		prepareMockUserRepoFn    func(m *mock_repository.MockUser)
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		wantErr                  bool
	}{
		{
			name:             "Spotifyとの同期が取れていることが確認されると、currentOperationがPlayの時はイベントは送信されない",
			sessionID:        "sessionID",
			currentOperation: "Play",
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
			prepareMockPusherFn:   func(m *mock_event.MockPusher) {},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: "PLAY",
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:06QTSGUEgcmKwiEJ0IMPig"},
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: false,
					ProgressWhenPaused:     0,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: "PLAY",
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:06QTSGUEgcmKwiEJ0IMPig"},
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: false,
					ProgressWhenPaused:     0,
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name:             "Spotifyとの同期が取れていることが確認されると、currentOperationがNextTrackの時はイベントが送信される",
			sessionID:        "sessionID",
			currentOperation: "NextTrack",
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
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: "PLAY",
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:06QTSGUEgcmKwiEJ0IMPig"},
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: false,
					ProgressWhenPaused:     0,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: "PLAY",
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:06QTSGUEgcmKwiEJ0IMPig"},
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: false,
					ProgressWhenPaused:     0,
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name:             "Spotifyとの同期が取れていることが確認されると、新しく追加すべき曲がSpotifyのキューに追加される",
			sessionID:        "sessionID",
			currentOperation: "NextTrack",
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
				m.EXPECT().Enqueue(gomock.Any(), "spotify:track:track3", "deviceID").Return(nil)
			},
			prepareMockPusherFn: func(m *mock_event.MockPusher) {
				m.EXPECT().Push(&event.PushMessage{
					SessionID: "sessionID",
					Msg:       entity.NewEventNextTrack(1),
				})
			},
			prepareMockUserRepoFn: func(m *mock_repository.MockUser) {},
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: "PLAY",
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:06QTSGUEgcmKwiEJ0IMPig"},
						{Index: 2, URI: "spotify:track:track2"},
						{Index: 3, URI: "spotify:track:track3"},
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: false,
					ProgressWhenPaused:     0,
				}, nil)
				m.EXPECT().Update(gomock.Any(), &entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: "PLAY",
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:06QTSGUEgcmKwiEJ0IMPig"},
						{Index: 2, URI: "spotify:track:track2"},
						{Index: 3, URI: "spotify:track:track3"},
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: false,
					ProgressWhenPaused:     0,
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name:             "Spotifyとの同期が取れていないとhandleInterruptが呼び出されErrorが返る",
			sessionID:        "sessionID",
			currentOperation: "NextTrack",
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
				m.EXPECT().FindByIDForUpdate(gomock.Any(), "sessionID").Return(&entity.Session{
					ID:        "sessionID",
					Name:      "name",
					CreatorID: "creatorID",
					DeviceID:  "deviceID",
					StateType: "PLAY",
					QueueHead: 1,
					QueueTracks: []*entity.QueueTrack{
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:hogehogehogehogehogeho"},
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: false,
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
						{Index: 0, URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
						{Index: 1, URI: "spotify:track:hogehogehogehogehogeho"},
					},
					ExpiredAt:              time.Time{},
					AllowToControlByOthers: false,
					ProgressWhenPaused:     0,
				}).Return(nil)
			},
			wantErr: true,
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

			tmpWaitTimeBeforeHandleTrackEnd := waitTimeAfterHandleTrackEnd
			waitTimeAfterHandleTrackEnd = 0
			defer func() {
				waitTimeAfterHandleTrackEnd = tmpWaitTimeBeforeHandleTrackEnd
			}()

			syncCheckTimerManager := entity.NewSyncCheckTimerManager()

			s := NewSessionTimerUseCase(mockSessionRepo, mockPlayer, mockPusher, syncCheckTimerManager)

			triggerAfterTrackEnd := s.tm.CreateExpiredTimer(tt.sessionID)
			if tt.currentOperation == operationNextTrack {
				triggerAfterTrackEnd.LockNextCh()
			}

			if _, err := s.handleWaitTimerExpiredTx(tt.sessionID, triggerAfterTrackEnd, tt.currentOperation)(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("handleWaitTimerExpired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
