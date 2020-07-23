package usecase

import (
	"context"
	"testing"

	"github.com/camphor-/relaym-server/domain/mock_spotify"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/mock_repository"

	"github.com/golang/mock/gomock"
)

func TestSessionUseCase_CanConnectToPusher(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		sessionID                string
		prepareMockSessionRepoFn func(m *mock_repository.MockSession)
		prepareMockPlayerFn      func(m *mock_spotify.MockPlayer)
		wantErr                  bool
	}{
		{
			name:      "存在しないセッションのとき404",
			sessionID: "not_found_session_id",
			prepareMockSessionRepoFn: func(m *mock_repository.MockSession) {
				m.EXPECT().FindByID("not_found_session_id").Return(nil, entity.ErrSessionNotFound)
			},
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			wantErr:             true,
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
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			//			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {
			//				m.EXPECT().CurrentlyPlaying(gomock.Any()).Return(&entity.CurrentPlayingInfo{
			//					Playing:  true,
			//					Progress: 10000000,
			//					Track: &entity.Track{
			//						URI:      "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
			//						ID:       "06QTSGUEgcmKwiEJ0IMPig",
			//						Name:     "Borderland",
			//						Duration: 213066000000,
			//						Artists:  []*entity.Artist{{Name: "MONOEYES"}},
			//						URL:      "https://open.spotify.com/track/06QTSGUEgcmKwiEJ0IMPig",
			//						Album: &entity.Album{
			//							Name: "Interstate 46 E.P.",
			//							Images: []*entity.AlbumImage{
			//								{
			//									URL:    "https://i.scdn.co/image/ab67616d0000b273b48630d6efcebca2596120c4",
			//									Height: 640,
			//									Width:  640,
			//								},
			//							},
			//						},
			//					},
			//				}, nil)
			//			},
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
			prepareMockPlayerFn: func(m *mock_spotify.MockPlayer) {},
			//				m.EXPECT().CurrentlyPlaying(gomock.Any()).Return(&entity.CurrentPlayingInfo{
			//					Playing:  true,
			//					Progress: 10000000,
			//					Track: &entity.Track{
			//						URI:      "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
			//						ID:       "06QTSGUEgcmKwiEJ0IMPig",
			//						Name:     "Borderland",
			//						Duration: 213066000000,
			//						Artists:  []*entity.Artist{{Name: "MONOEYES"}},
			//						URL:      "https://open.spotify.com/track/06QTSGUEgcmKwiEJ0IMPig",
			//						Album: &entity.Album{
			//							Name: "Interstate 46 E.P.",
			//							Images: []*entity.AlbumImage{
			//								{
			//									URL:    "https://i.scdn.co/image/ab67616d0000b273b48630d6efcebca2596120c4",
			//									Height: 640,
			//									Width:  640,
			//								},
			//							},
			//						},
			//					},
			//				}, nil)
			//			},
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
			mockPlayer := mock_spotify.NewMockPlayer(ctrl)
			tt.prepareMockPlayerFn(mockPlayer)
			syncCheckTimerManager := entity.NewSyncCheckTimerManager()
			stUC := NewSessionTimerUseCase(nil, mockPlayer, nil, syncCheckTimerManager)
			s := NewSessionUseCase(mockSessionRepo, nil, mockPlayer, nil, nil, nil, stUC)

			if err := s.CanConnectToPusher(context.Background(), tt.sessionID); (err != nil) != tt.wantErr {
				t.Errorf("CanConnectToPusher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
