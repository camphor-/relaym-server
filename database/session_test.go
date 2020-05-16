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
	queueTrack := &queueTrackDTO{
		Index:     0,
		URI:       "existing_uri",
		SessionID: "existing_session_id",
	}
	if err := dbMap.Insert(user, session, queueTrack); err != nil {
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
				QueueTracks: []*entity.QueueTrack{
					{
						Index:     0,
						URI:       "existing_uri",
						SessionID: "existing_session_id",
					},
				},
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

func TestSessionRepository_Update(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(sessionDTO{}, "sessions")
	dbMap.AddTableWithName(userDTO{}, "users")
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
		StateType: "PAUSE",
	}
	if err := dbMap.Insert(user, session); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		session *entity.Session
		wantErr bool
	}{
		{
			name: "既に存在するセッションの情報を更新できる",
			session: &entity.Session{
				ID:          "existing_session_id",
				Name:        "existing_session_name",
				CreatorID:   "existing_user",
				QueueHead:   1,
				StateType:   entity.Play,
				QueueTracks: []*entity.QueueTrack{},
			},
			wantErr: false,
		},
		{
			name: "存在しないセッションはエラーになる",
			session: &entity.Session{
				ID:          "not_found_id",
				Name:        "not_found_id_names",
				CreatorID:   "existing_user",
				QueueHead:   1,
				StateType:   entity.Play,
				QueueTracks: []*entity.QueueTrack{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewSessionRepository(dbMap)
			if err := r.Update(tt.session); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got, err := r.FindByID(tt.session.ID)
				if err != nil {
					t.Fatal(err)
				}

				if !cmp.Equal(tt.session, got) {
					t.Errorf("Update() diff = %v", cmp.Diff(got, tt.session))

				}
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
	sessionHasQueueTrack := &sessionDTO{
		ID:        "session_with_queue_track_id",
		Name:      "session_with_queue_track_name",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
	}
	sessionHasNoQueueTrack := &sessionDTO{
		ID:        "session_with_no_queue_track_id",
		Name:      "session_with_no_queue_track_name",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
	}
	queueTracks := &queueTrackDTO{
		Index:     0,
		URI:       "uri",
		SessionID: "session_with_queue_track_id",
	}
	if err := dbMap.Insert(user, sessionHasQueueTrack, sessionHasNoQueueTrack, queueTracks); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		queueTrack *entity.QueueTrackToStore
		wantIndex  int
		wantErr    error
	}{
		{
			name: "すでにひも付いているqueue_tracksが1つ以上存在するsessionsに新規queue_tracksを正しく紐づけて保存できる",
			queueTrack: &entity.QueueTrackToStore{
				URI:       "new_uri",
				SessionID: "session_with_queue_track_id",
			},
			wantIndex: 1,
			wantErr:   nil,
		},
		{
			name: "ひも付いているqueue_tracksが1つも存在しないsessionsに新規queue_tracksを正しく紐づけて保存できる",
			queueTrack: &entity.QueueTrackToStore{
				URI:       "new_uri",
				SessionID: "session_with_no_queue_track_id",
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
			if err := r.StoreQueueTrack(tt.queueTrack); !errors.Is(err, tt.wantErr) {
				t.Errorf("SessionRepository.StoreQueueTracks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				queueTracks, _ := r.getQueueTracksBySessionID(tt.queueTrack.SessionID)
				queueTrack, notFound := findQueueTrackByIndexAndSessionID(queueTracks, tt.wantIndex, tt.queueTrack.SessionID)

				if (notFound != nil) || (queueTrack.URI != tt.queueTrack.URI) {
					t.Errorf("SessionRepository.StoreQueueTrack() queue_track not found. wantIndex %v, wantSessionID %v", tt.wantIndex, tt.queueTrack.SessionID)
				}
			}
		})
	}
}

func TestSessionRepository_getQueueTrackBySessionID(t *testing.T) {
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
	sessionHasManyQueueTracks := &sessionDTO{
		ID:        "session_has_many_queue_tracks_id",
		Name:      "session_has_many_queue_tracks_name",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
	}

	queueTrack1 := &queueTrackDTO{
		Index:     0,
		URI:       "existing_uri1",
		SessionID: "existing_session_id",
	}
	queueTrack2 := &queueTrackDTO{
		Index:     0,
		URI:       "existing_uri2",
		SessionID: "session_has_many_queue_tracks_id",
	}
	queueTrack3 := &queueTrackDTO{
		Index:     1,
		URI:       "existing_uri3",
		SessionID: "session_has_many_queue_tracks_id",
	}
	queueTrack4 := &queueTrackDTO{
		Index:     2,
		URI:       "existing_uri4",
		SessionID: "session_has_many_queue_tracks_id",
	}
	if err := dbMap.Insert(user, session, sessionHasManyQueueTracks, queueTrack1, queueTrack2, queueTrack3, queueTrack4); err != nil {
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
					URI:       "existing_uri1",
					SessionID: "existing_session_id",
				},
			},
			wantErr: nil,
		},
		{
			name: "セッションIDから正しく複数のqueue_tracksを正しい順序で取り出せる",
			id:   "session_has_many_queue_tracks_id",
			want: []*entity.QueueTrack{
				{
					Index:     0,
					URI:       "existing_uri2",
					SessionID: "session_has_many_queue_tracks_id",
				},
				{
					Index:     1,
					URI:       "existing_uri3",
					SessionID: "session_has_many_queue_tracks_id",
				},
				{
					Index:     2,
					URI:       "existing_uri4",
					SessionID: "session_has_many_queue_tracks_id",
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
			queueTracks, err := r.getQueueTracksBySessionID(tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SessionRepository.getQueueTrackBySessionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(queueTracks, tt.want) {
				t.Errorf("SessionRepository.getQueueTracksBySessionID() diff = %v", cmp.Diff(queueTracks, tt.want))
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
