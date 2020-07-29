package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"

	"golang.org/x/oauth2"

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
		ID:                     "existing_session_id",
		Name:                   "existing_session_name",
		CreatorID:              "existing_user",
		QueueHead:              0,
		StateType:              "PLAY",
		DeviceID:               "device_id",
		ExpiredAt:              time.Date(2020, time.December, 1, 12, 0, 0, 0, time.UTC),
		AllowToControlByOthers: true,
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
				DeviceID:  "device_id",
				StateType: "PLAY",
				QueueHead: 0,
				QueueTracks: []*entity.QueueTrack{
					{
						Index:     0,
						URI:       "existing_uri",
						SessionID: "existing_session_id",
					},
				},
				ExpiredAt:              time.Date(2020, time.December, 1, 12, 0, 0, 0, time.UTC),
				AllowToControlByOthers: true,
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
			got, err := r.FindByID(context.TODO(), tt.id)
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

func TestSessionRepository_FindByIDForUpdate(t *testing.T) {
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
		ID:                     "existing_session_id",
		Name:                   "existing_session_name",
		CreatorID:              "existing_user",
		QueueHead:              0,
		StateType:              "PLAY",
		DeviceID:               "device_id",
		ExpiredAt:              time.Date(2020, time.December, 1, 12, 0, 0, 0, time.UTC),
		AllowToControlByOthers: true,
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
				DeviceID:  "device_id",
				StateType: "PLAY",
				QueueHead: 0,
				QueueTracks: []*entity.QueueTrack{
					{
						Index:     0,
						URI:       "existing_uri",
						SessionID: "existing_session_id",
					},
				},
				ExpiredAt:              time.Date(2020, time.December, 1, 12, 0, 0, 0, time.UTC),
				AllowToControlByOthers: true,
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
			got, err := r.FindByIDForUpdate(context.TODO(), tt.id)
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
		ID:                     "existing_session_id",
		Name:                   "existing_session_name",
		CreatorID:              "existing_user",
		QueueHead:              0,
		StateType:              "PLAY",
		DeviceID:               "device_id",
		ExpiredAt:              time.Now(),
		AllowToControlByOthers: true,
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
				ID:                     "new_session_id",
				Name:                   "new_session_name",
				CreatorID:              "existing_user",
				DeviceID:               "new_device_id",
				StateType:              "PLAY",
				QueueHead:              0,
				QueueTracks:            nil,
				AllowToControlByOthers: true,
			},
			wantErr: nil,
		},
		{
			name: "登録済みのsessionの場合ErrSessionAlreadyExistedを返す",
			session: &entity.Session{
				ID:                     "existing_session_id",
				Name:                   "existing_session_name",
				CreatorID:              "existing_user",
				DeviceID:               "device_id",
				StateType:              "PLAY",
				QueueHead:              0,
				QueueTracks:            nil,
				AllowToControlByOthers: true,
			},
			wantErr: entity.ErrSessionAlreadyExisted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SessionRepository{
				dbMap: dbMap,
			}
			if err := r.StoreSession(context.TODO(), tt.session); !errors.Is(err, tt.wantErr) {
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
		ID:                     "existing_session_id",
		Name:                   "existing_session_name",
		CreatorID:              "existing_user",
		QueueHead:              0,
		StateType:              "PAUSE",
		DeviceID:               "device_id",
		ExpiredAt:              time.Date(2020, time.December, 1, 12, 0, 0, 0, time.UTC),
		AllowToControlByOthers: false,
	}
	sameFieldSession := &sessionDTO{
		ID:                     "same_field_session_id",
		Name:                   "same_field_session_name",
		CreatorID:              "existing_user",
		QueueHead:              0,
		StateType:              "PAUSE",
		ExpiredAt:              time.Date(2020, time.December, 1, 12, 0, 0, 0, time.UTC),
		AllowToControlByOthers: true,
	}
	if err := dbMap.Insert(user, session, sameFieldSession); err != nil {
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
				ID:                     "existing_session_id",
				Name:                   "existing_session_name",
				CreatorID:              "existing_user",
				DeviceID:               "new_device_id",
				StateType:              entity.Play,
				QueueHead:              1,
				QueueTracks:            []*entity.QueueTrack{},
				ExpiredAt:              time.Date(2020, time.December, 1, 12, 0, 0, 0, time.UTC),
				AllowToControlByOthers: true,
			},
			wantErr: false,
		},
		{
			name: "フィールドの値が全てDBの値を一致するセッションで更新してもエラーにならない",
			session: &entity.Session{
				ID:                     "same_field_session_id",
				Name:                   "same_field_session_name",
				CreatorID:              "existing_user",
				QueueHead:              0,
				StateType:              entity.Pause,
				QueueTracks:            []*entity.QueueTrack{},
				ExpiredAt:              time.Date(2020, time.December, 1, 12, 0, 0, 0, time.UTC),
				AllowToControlByOthers: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewSessionRepository(dbMap)
			if err := r.Update(context.TODO(), tt.session); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := r.FindByID(context.TODO(), tt.session.ID)
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
		ID:                     "session_with_queue_track_id",
		Name:                   "session_with_queue_track_name",
		CreatorID:              "existing_user",
		QueueHead:              0,
		StateType:              "PLAY",
		ExpiredAt:              time.Now(),
		AllowToControlByOthers: true,
	}
	sessionHasNoQueueTrack := &sessionDTO{
		ID:                     "session_with_no_queue_track_id",
		Name:                   "session_with_no_queue_track_name",
		CreatorID:              "existing_user",
		QueueHead:              0,
		StateType:              "PLAY",
		ExpiredAt:              time.Now(),
		AllowToControlByOthers: true,
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
			if err := r.StoreQueueTrack(context.TODO(), tt.queueTrack); !errors.Is(err, tt.wantErr) {
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
		ExpiredAt: time.Now(),
	}
	sessionHasManyQueueTracks := &sessionDTO{
		ID:        "session_has_many_queue_tracks_id",
		Name:      "session_has_many_queue_tracks_name",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
		ExpiredAt: time.Now(),
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

func TestSessionRepository_FindCreatorTokenBySessionID(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(spotifyAuthDTO{}, "spotify_auth")
	dbMap.AddTableWithName(userDTO{}, "users")
	dbMap.AddTableWithName(sessionDTO{}, "sessions")
	truncateTable(t, dbMap)
	if err := dbMap.Insert(&userDTO{ID: "creator_user_id", SpotifyUserID: "new_user_spotify"}); err != nil {
		t.Fatal(err)
	}
	if err := dbMap.Insert(&spotifyAuthDTO{
		UserID:       "creator_user_id",
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		Expiry:       time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
	if err := dbMap.Insert(&sessionDTO{
		ID:        "exist_session_id",
		Name:      "session_name",
		CreatorID: "creator_user_id",
		QueueHead: 0,
		StateType: "STOP",
		DeviceID:  "device_id",
		ExpiredAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		sessionID     string
		wantToken     *oauth2.Token
		wantCreatorID string
		wantErr       error
	}{
		{
			name:      "正常系",
			sessionID: "exist_session_id",
			wantToken: &oauth2.Token{
				AccessToken:  "access_token",
				TokenType:    "Bearer",
				RefreshToken: "refresh_token",
				Expiry:       time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			wantCreatorID: "creator_user_id",
			wantErr:       nil,
		},
		{
			name:          "sessionが存在しないとErrSessionNotFound",
			sessionID:     "not_exist_session_id",
			wantToken:     nil,
			wantCreatorID: "",
			wantErr:       entity.ErrSessionNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SessionRepository{
				dbMap: dbMap,
			}
			token, creatorID, err := r.FindCreatorTokenBySessionID(context.TODO(), tt.sessionID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SessionRepository.FindCreatorTokenBySessionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			opt := cmpopts.IgnoreUnexported(oauth2.Token{})
			if !cmp.Equal(token, tt.wantToken, opt) {
				t.Errorf("SessionRepository.FindCreatorTokenBySessionID() diff = %v", cmp.Diff(token, tt.wantToken))
				return
			}

			if creatorID != tt.wantCreatorID {
				t.Errorf("SessionRepository.FindCreatorTokenBySessionID() diff = %v", cmp.Diff(creatorID, tt.wantCreatorID))
				return
			}
		})
	}
}

func TestSessionRepository_ArchiveSessionsForBatch(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	user := &userDTO{
		ID:            "existing_user",
		SpotifyUserID: "existing_user_spotify",
		DisplayName:   "existing_user_display_name",
	}
	oldSession := &sessionDTO{
		ID:        "existing_session_id1",
		Name:      "existing_session_name",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
		DeviceID:  "device_id",
		ExpiredAt: time.Now().Add(-1 * 24 * time.Hour),
	}

	newSession := &sessionDTO{
		ID:        "existing_session_id2",
		Name:      "existing_session_name",
		CreatorID: "existing_user",
		QueueHead: 0,
		StateType: "PLAY",
		DeviceID:  "device_id",
		ExpiredAt: time.Now().Add(1 * 24 * time.Hour),
	}

	tests := []struct {
		name      string
		session   *sessionDTO
		wantState entity.StateType
	}{
		{
			name:      "新しいsessionはARCHIVEされない",
			session:   newSession,
			wantState: "PLAY",
		},
		{
			name:      "古いsessionはARCHIVEされる",
			session:   oldSession,
			wantState: "ARCHIVED",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMap.AddTableWithName(userDTO{}, "users")
			dbMap.AddTableWithName(sessionDTO{}, "sessions")
			truncateTable(t, dbMap)

			if err := dbMap.Insert(user); err != nil {
				t.Errorf("SessionRepository.ArchiveSessionsForBatch() error = %v", err)
				return
			}

			if err := dbMap.Insert(tt.session); err != nil {
				t.Errorf("SessionRepository.ArchiveSessionsForBatch() error = %v", err)
				return
			}

			r := &SessionRepository{
				dbMap: dbMap,
			}
			if err := r.ArchiveSessionsForBatch(); err != nil {
				t.Errorf("SessionRepository.ArchiveSessionsForBatch() error = %v", err)
				return
			}

			session, err := r.FindByID(context.TODO(), tt.session.ID)
			if err != nil {
				t.Errorf("SessionRepository.ArchiveSessionsForBatch() error = %v", err)
				return
			}

			if session.StateType != tt.wantState {
				t.Errorf("SessionRepository.ArchiveSessionsForBatch() wantState = %s, state = %s", session.StateType, tt.wantState)
				return
			}
		})
	}
}

func TestSessionRepository_UpdateWithExpiredAt(t *testing.T) {
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
		ID:                     "existing_session_id",
		Name:                   "existing_session_name",
		CreatorID:              "existing_user",
		QueueHead:              0,
		StateType:              "PAUSE",
		DeviceID:               "device_id",
		ExpiredAt:              time.Date(2020, time.December, 1, 12, 0, 0, 0, time.UTC),
		AllowToControlByOthers: true,
	}
	if err := dbMap.Insert(user, session); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		session      *entity.Session
		newExpiredAt time.Time
		wantErr      bool
	}{
		{
			name: "正常に動作し、expiredAtも更新される",
			session: &entity.Session{
				ID:                     "existing_session_id",
				Name:                   "existing_session_name",
				CreatorID:              "existing_user",
				DeviceID:               "new_device_id",
				StateType:              entity.Play,
				QueueHead:              1,
				QueueTracks:            []*entity.QueueTrack{},
				AllowToControlByOthers: true,
			},
			newExpiredAt: time.Date(2020, time.December, 7, 12, 0, 0, 0, time.UTC),
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewSessionRepository(dbMap)
			if err := r.UpdateWithExpiredAt(context.TODO(), tt.session, tt.newExpiredAt); (err != nil) != tt.wantErr {
				t.Errorf("UpdateWithExpiredAt() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				var dto sessionDTO

				if err := r.dbMap.SelectOne(&dto, "SELECT id, name, creator_id, queue_head, state_type, device_id, expired_at FROM sessions WHERE id = ?", tt.session.ID); err != nil {
					t.Fatal(err)
				}

				if dto.ExpiredAt != tt.newExpiredAt {
					t.Errorf("UpdateWithExpiredAt() want expierdAt: %s, got: %s", tt.newExpiredAt, dto.ExpiredAt)
				}
			}
		})
	}
}
