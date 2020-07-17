package usecase

import (
	"context"
	"testing"

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
			stUC := NewSessionTimerUseCase(nil, nil, nil)
			s := NewSessionUseCase(mockSessionRepo, nil, nil, nil, nil, nil, stUC)

			if err := s.CanConnectToPusher(context.Background(), tt.sessionID); (err != nil) != tt.wantErr {
				t.Errorf("CanConnectToPusher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
