package entity

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewSession(t *testing.T) {
	t.Parallel()

	session := &Session{
		ID:                     "ID",
		Name:                   "VeryGoodSession",
		CreatorID:              "VeryCreativePersonID",
		StateType:              Stop,
		QueueHead:              0,
		QueueTracks:            nil,
		AllowToControlByOthers: true,
	}

	tests := []struct {
		name                   string
		sessionName            string
		creatorID              string
		allowToCOntrolByOthers bool
		want                   *Session
	}{
		{
			name:                   "正常系",
			sessionName:            "VeryGoodSession",
			creatorID:              "VeryCreativePersonID",
			allowToCOntrolByOthers: true,
			want:                   session,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSession(tt.sessionName, tt.creatorID, tt.allowToCOntrolByOthers)
			if err != nil {
				t.Fatal(err)
			}
			opts := []cmp.Option{cmpopts.IgnoreFields(Session{}, "ID"), cmpopts.IgnoreFields(Session{}, "ExpiredAt")}
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
			name: "作成者でないときtrue",
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
			wantErr: false,
		},
		{
			name: "Archived",
			session: &Session{
				StateType: Archived,
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
			name: "Stopかつ次に再生するトラックが存在する",
			session: &Session{
				StateType:   Stop,
				QueueHead:   0,
				QueueTracks: []*QueueTrack{{}},
			},
			wantErr: false,
		},
		{
			name: "Stopかつ次に再生するトラックが存在しない",
			session: &Session{
				StateType:   Stop,
				QueueHead:   0,
				QueueTracks: nil,
			},
			wantErr: true,
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

func TestSession_GoNextTrack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		s       *Session
		wantErr bool
	}{
		{
			name: "一つも曲が追加されてないときはエラー",
			s: &Session{
				QueueHead:   0,
				QueueTracks: nil,
			},
			wantErr: true,
		},
		{
			name: "最後の曲を再生していたときはエラー",
			s: &Session{
				QueueHead: 2,
				QueueTracks: []*QueueTrack{
					{},
					{},
					{}, // 再生中
				},
			},
			wantErr: true,
		},
		{
			name: "次の曲が存在するときはエラーにならない",
			s: &Session{
				QueueHead: 2,
				QueueTracks: []*QueueTrack{
					{},
					{},
					{}, // 再生中
					{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.GoNextTrack(); (err != nil) != tt.wantErr {
				t.Errorf("GoNextTrack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSession_TrackURIsShouldBeAddedWhenStopToPlay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		s       *Session
		want    []string
		wantErr bool
	}{
		{
			name: "まだキューに何も追加していないときエラー",
			s: &Session{
				QueueTracks: []*QueueTrack{},
				QueueHead:   0,
				StateType:   Stop,
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "キューに1曲追加して、まだ再生を始めていないときは長さ1のスライス",
			s: &Session{
				QueueTracks: []*QueueTrack{{URI: "0"}},
				QueueHead:   0,
				StateType:   Stop,
			},
			want:    []string{"0"},
			wantErr: false,
		},
		{
			name: "キューが1曲で1曲目を再生終了してSTOPになったときは、再生する曲がないのでエラー",
			s: &Session{
				QueueTracks: []*QueueTrack{{URI: "0"}},
				QueueHead:   1,
				StateType:   Stop,
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "キューに2曲追加して、まだ再生を始めていないときは長さ2のスライス",
			s: &Session{
				QueueTracks: []*QueueTrack{{URI: "0"}, {URI: "1"}},
				QueueHead:   0,
				StateType:   Stop,
			},
			want:    []string{"0", "1"},
			wantErr: false,
		},
		{
			name: "キューに4曲追加して、まだ再生を始めていない時は長さ3のスライス",
			s: &Session{
				QueueTracks: []*QueueTrack{{URI: "0"}, {URI: "1"}, {URI: "2"}, {URI: "3"}},
				QueueHead:   0,
				StateType:   Stop,
			},
			want:    []string{"0", "1", "2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.s.TrackURIsShouldBeAddedWhenStopToPlay()
			if (err != nil) != tt.wantErr {
				t.Errorf("TrackURIsShouldBeAddedWhenStartPlay() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("TrackURIsShouldBeAddedWhenStartPlay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_TrackURIShouldBeAddedWhenHandleTrackEnd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    *Session
		want string
	}{
		{
			name: "正常に二曲先のTrackのURIが返る",
			s: &Session{
				QueueTracks: []*QueueTrack{{URI: "0"}, {URI: "1"}, {URI: "2"}, {URI: "3"}},
				QueueHead:   0,
				StateType:   Play,
			},
			want: "2",
		},
		{
			name: "二曲先の曲が存在しない時には空文字列が返る",
			s: &Session{
				QueueTracks: []*QueueTrack{{URI: "0"}, {URI: "1"}, {URI: "2"}, {URI: "3"}},
				QueueHead:   2,
				StateType:   Play,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := tt.s.TrackURIShouldBeAddedWhenHandleTrackEnd()

			if !cmp.Equal(got, tt.want) {
				t.Errorf("TrackURIShouldBeAddedWhenHandleTrackEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_IsPlayingCorrectTrack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		session     *Session
		playingInfo *CurrentPlayingInfo
		wantErr     bool
	}{
		{
			name: "キューが空で何も再生していないときはエラーにならない",
			session: &Session{
				StateType:   Stop,
				QueueHead:   0,
				QueueTracks: nil,
			},
			playingInfo: nil,
			wantErr:     false,
		},
		{
			name: "キューが空でなにか再生していてもはエラーにならない",
			session: &Session{
				StateType:   Stop,
				QueueHead:   0,
				QueueTracks: nil,
			},
			playingInfo: &CurrentPlayingInfo{
				Playing:  true,
				Progress: 0,
				Track: &Track{
					URI: "spotify:track:hoge",
				},
				Device: nil,
			},
			wantErr: false,
		},
		{
			name: "再生されている曲が実際に再生されていればエラーにならない",
			session: &Session{
				StateType: Play,
				QueueHead: 0,
				QueueTracks: []*QueueTrack{
					{URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
				},
			},
			playingInfo: &CurrentPlayingInfo{
				Playing:  true,
				Progress: 0,
				Track:    &Track{URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
				Device:   nil,
			},
			wantErr: false,
		},
		{
			name: "再生されているはずなのに再生されていなければエラー",
			session: &Session{
				StateType: Play,
				QueueHead: 0,
				QueueTracks: []*QueueTrack{
					{URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
				},
			},
			playingInfo: nil,
			wantErr:     true,
		},
		{
			name: "再生されているはずなのにSpotify側が一時停止していたらエラー",
			session: &Session{
				StateType: Play,
				QueueHead: 0,
				QueueTracks: []*QueueTrack{
					{URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
				},
			},
			playingInfo: &CurrentPlayingInfo{
				Playing:  false,
				Progress: 0,
				Track:    &Track{},
				Device:   &Device{},
			},
			wantErr: true,
		},
		{
			name: "再生が終了してStopになっていたらエラーにならない",
			session: &Session{
				StateType: Stop,
				QueueHead: 1,
				QueueTracks: []*QueueTrack{
					{URI: "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"},
				},
			},
			playingInfo: nil,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tt.session.IsPlayingCorrectTrack(tt.playingInfo); (err != nil) != tt.wantErr {
				t.Errorf("IsPlayingCorrectTrack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
