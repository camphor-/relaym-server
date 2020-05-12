package database

import (
	"errors"
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/google/go-cmp/cmp"
)

func TestSessionRepository_FindByID(t *testing.T) {
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(sessionDTO{}, "sessions")
	dbMap.AddTableWithName(userDTO{}, "users")
	truncateTable(t, dbMap)
	user := &userDTO{
		ID:            "existing_user",
		SpotifyUserID: "existing_user_spotify",
		DisplayName:   "existing_user_display_name",
	}
	session := &sessionDTO{
		ID:        "existing_session_id",
		Name:      "existing_session_name",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
	}
	if err := dbMap.Insert(user, session); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		id      string
		want    *entity.Session
		wantErr error
	}{
		{
			name: "存在するsessionを正しく取得できる",
			id:   "existing_session_id",
			want: &entity.Session{
				ID:        "existing_session_id",
				Name:      "existing_session_name",
				CreatorID: "existing_user",
				QueueHead: 0,
				StateType: "PLAY",
			},
			wantErr: nil,
		},
		{
			name:    "存在しないidの場合はErrSessionNotFound",
			id:      "not_exist_session_id",
			want:    nil,
			wantErr: entity.ErrSessionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SessionRepository{dbMap: dbMap}
			got, err := r.FindByID(tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SessionRepository.FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("SessionRepository.FindByID() diff=%v", cmp.Diff(tt.want, got))
				return
			}
		})
	}
}

func TestSessionRepository_StoreSession(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(sessionDTO{}, "sessions")
	dbMap.AddTableWithName(userDTO{}, "users")
	truncateTable(t, dbMap)
	user := &userDTO{
		ID:            "existing_user",
		SpotifyUserID: "existing_user_spotify",
		DisplayName:   "existing_user_display_name",
	}
	session := &sessionDTO{
		ID:        "existing_session_id",
		Name:      "existing_session_name",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
	}
	if err := dbMap.Insert(user, session); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		session *entity.Session
		wantErr error
	}{
		{
			name: "新規sessionを正しく保存できる",
			session: &entity.Session{
				ID:        "new_session_id",
				Name:      "new_session_name",
				CreatorID: "existing_user",
				QueueHead: 0,
				StateType: "PLAY",
			},
			wantErr: nil,
		},
		{
			name: "登録済みのsessionの場合ErrSessionAlreadyExistedを返す",
			session: &entity.Session{
				ID:        "existing_session_id",
				Name:      "existing_session_name",
				CreatorID: "existing_user",
				QueueHead: 0,
				StateType: "PLAY",
			},
			wantErr: entity.ErrSessionAlreadyExisted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SessionRepository{
				dbMap: dbMap,
			}
			if err := r.StoreSession(tt.session); !errors.Is(err, tt.wantErr) {
				t.Errorf("SessionRepository.StoreSessions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSessionRepository_StoreQueueTrack(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(userDTO{}, "users")
	dbMap.AddTableWithName(sessionDTO{}, "sessions")
	dbMap.AddTableWithName(queueTrackDTO{}, "queue_tracks")
	truncateTable(t, dbMap)
	user := &userDTO{
		ID:            "existing_user",
		SpotifyUserID: "existing_user_spotify",
		DisplayName:   "existing_user_display_name",
	}
	session_a := &sessionDTO{
		ID:        "existing_session_id_a",
		Name:      "existing_session_name_a",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
	}
	session_b := &sessionDTO{
		ID:        "existing_session_id_b",
		Name:      "existing_session_name_b",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
	}
	queue_tracks := &queueTrackDTO{
		Index:     0,
		URI:       "existing_uri",
		SessionID: "existing_session_id_a",
	}
	if err := dbMap.Insert(user, session_a, session_b, queue_tracks); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		queue_track *entity.QueueTrackToStore
		wantIndex   int
		wantErr     error
	}{
		{
			name: "すでにひも付いているqueue_tracksが1つ以上存在するsessionsに新規queue_tracksを正しく紐づけて保存できる",
			queue_track: &entity.QueueTrackToStore{
				URI:       "new_uri",
				SessionID: "existing_session_id_a",
			},
			wantIndex: 1,
			wantErr:   nil,
		},
		{
			name: "ひも付いているqueue_tracksが1つも存在しないsessionsに新規queue_tracksを正しく紐づけて保存できる",
			queue_track: &entity.QueueTrackToStore{
				URI:       "new_uri",
				SessionID: "existing_session_id_b",
			},
			wantIndex: 0,
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SessionRepository{
				dbMap: dbMap,
			}
			if err := r.StoreQueueTrack(tt.queue_track); !errors.Is(err, tt.wantErr) {
				t.Errorf("SessionRepository.StoreQueueTracks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				queue_tracks, _ := r.GetQueueTracksBySessionID(tt.queue_track.SessionID)
				queue_track, notFound := findQueueTrackByIndexAndSessionID(queue_tracks, tt.wantIndex, tt.queue_track.SessionID)

				if (notFound != nil) || (queue_track.URI != tt.queue_track.URI) {
					t.Errorf("SessionRepository.StoreQueueTrack() queue_track not found. wantIndex %v, wantSessionID %v", tt.wantIndex, tt.queue_track.SessionID)
				}
			}
		})
	}
}

func TestSessionRepository_GetQueueTrackBySessionID(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(userDTO{}, "users")
	dbMap.AddTableWithName(sessionDTO{}, "sessions")
	dbMap.AddTableWithName(queueTrackDTO{}, "queue_tracks")
	truncateTable(t, dbMap)
	user := &userDTO{
		ID:            "existing_user",
		SpotifyUserID: "existing_user_spotify",
		DisplayName:   "existing_user_display_name",
	}
	session := &sessionDTO{
		ID:        "existing_session_id",
		Name:      "existing_session_name",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
	}
	queue_tracks := &queueTrackDTO{
		Index:     0,
		URI:       "existing_uri",
		SessionID: "existing_session_id",
	}
	if err := dbMap.Insert(user, session, queue_tracks); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		id      string
		want    []*entity.QueueTrack
		wantErr error
	}{
		{
			name: "セッションIDから正しくqueue_tracksを取り出せる",
			id:   "existing_session_id",
			want: []*entity.QueueTrack{
				{
					Index:     0,
					URI:       "existing_uri",
					SessionID: "existing_session_id",
				},
			},
			wantErr: nil,
		},
		{
			name:    "存在しないセッションIDを渡すと空の[]*entity.QueueTrackが返ってくる",
			id:      "not_exist_session_id",
			want:    []*entity.QueueTrack{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SessionRepository{
				dbMap: dbMap,
			}
			queue_tracks, err := r.GetQueueTracksBySessionID(tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SessionRepository.GetQueueTrackBySessionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(queue_tracks, tt.want) {
				t.Errorf("SessionRepository.GetQueueTracksBySessionID() diff = %v", cmp.Diff(queue_tracks, tt.want))
			}
		})
	}
}

func findQueueTrackByIndexAndSessionID(queueTracks []*entity.QueueTrack, index int, sessionID string) (*entity.QueueTrack, error) {
	for _, qt := range queueTracks {
		if (qt.Index == index) && (qt.SessionID == sessionID) {
			return qt, nil
		}
	}
	return nil, errors.New("Not Found")
}
