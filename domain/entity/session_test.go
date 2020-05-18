package entity

import "testing"

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

func TestSession_MoveToPause(t *testing.T) {
	tests := []struct {
		name    string
		session *Session
		wantErr bool
	}{
		{
			name: "Play",
			session: &Session{
				StateType: Play,
			},
			wantErr: false,
		},
		{
			name: "Pause",
			session: &Session{
				StateType: Pause,
			},
			wantErr: false,
		},
		{
			name: "Stop",
			session: &Session{
				StateType: Stop,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.session.MoveToPause(); (err != nil) != tt.wantErr {
				t.Errorf("MoveToPause() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSession_MoveToPlay(t *testing.T) {
	tests := []struct {
		name    string
		session *Session
		wantErr bool
	}{
		{
			name: "Play",
			session: &Session{
				StateType: Play,
			},
			wantErr: false,
		},
		{
			name: "Pause",
			session: &Session{
				StateType: Pause,
			},
			wantErr: false,
		},
		{
			name: "Stop",
			session: &Session{
				StateType: Stop,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.session.MoveToPlay(); (err != nil) != tt.wantErr {
				t.Errorf("MoveToPlay() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
