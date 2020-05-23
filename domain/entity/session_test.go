package entity

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewSession(t *testing.T) {
	t.Parallel()

	session := &Session{
		ID:          "ID",
		Name:        "VeryGoodSession",
		CreatorID:   "VeryCreativePersonID",
		StateType:   Stop,
		QueueHead:   0,
		QueueTracks: nil,
	}

	tests := []struct {
		name        string
		sessionName string
		creatorID   string
		want        *Session
	}{
		{
			name:        "正常系",
			sessionName: "VeryGoodSession",
			creatorID:   "VeryCreativePersonID",
			want:        session,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSession(tt.sessionName, tt.creatorID)
			if err != nil {
				t.Fatal(err)
			}
			opts := []cmp.Option{cmpopts.IgnoreFields(Session{}, "ID")}
			if !cmp.Equal(got, tt.want, opts...) {
				t.Errorf("NewSession() diff = %v", cmp.Diff(got, tt.want, opts...))
			}
		})
	}
}

func TestNewStateType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stateType string
		want      StateType
		wantErr   bool
	}{
		{
			name:      "Play",
			stateType: "PLAY",
			want:      Play,
			wantErr:   false,
		},
		{
			name:      "Pause",
			stateType: "PAUSE",
			want:      Pause,
			wantErr:   false,
		},
		{
			name:      "Stop",
			stateType: "STOP",
			want:      Stop,
			wantErr:   false,
		},
		{
			name:      "無効なstate type",
			stateType: "invalid",
			want:      "",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStateType(tt.stateType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStateType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewStateType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		st   StateType
		want string
	}{
		{
			name: "Play",
			st:   Play,
			want: "PLAY",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.st.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_IsCreator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		session *Session
		userID  string
		want    bool
	}{
		{
			name: "作成者のときtrue",
			session: &Session{
				CreatorID: "user_id",
			},
			userID: "user_id",
			want:   true,
		},
		{
			name: "作成者でないときときtrue",
			session: &Session{
				CreatorID: "user_id",
			},
			userID: "not_creator_user_id",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.session.IsCreator(tt.userID); got != tt.want {
				t.Errorf("IsCreator() = %v, want %v", got, tt.want)
			}
		})
	}
}
