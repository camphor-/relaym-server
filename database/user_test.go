package database

import (
	"errors"
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/google/go-cmp/cmp"
)

func TestUserRepository_FindByID(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(userDTO{}, "users")
	truncateTable(t, dbMap)
	if err := dbMap.Insert(&userDTO{
		ID:            "existing_user",
		SpotifyUserID: "existing_user_spotify",
		DisplayName:   "display_name",
		DeviceID:      "device_id",
	}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		id      string
		want    *entity.User
		wantErr error
	}{
		{
			name: "存在するユーザを正しく取得できる",
			id:   "existing_user",
			want: &entity.User{
				ID:            "existing_user",
				SpotifyUserID: "existing_user_spotify",
				DisplayName:   "display_name",
				DeviceID:      "device_id",
			},
			wantErr: nil,
		},
		{
			name:    "存在しないspotifyUserIDの場合ErrUserNotFound",
			id:      "not_found",
			want:    nil,
			wantErr: entity.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &UserRepository{dbMap: dbMap}
			got, err := r.FindByID(tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("FindByID() diff=%v", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestUserRepository_FindBySpotifyUserID(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(userDTO{}, "users")
	truncateTable(t, dbMap)
	if err := dbMap.Insert(&userDTO{
		ID:            "existing_user",
		SpotifyUserID: "existing_user_spotify",
		DisplayName:   "display_name",
		DeviceID:      "device_id",
	}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		spotifyUserID string
		want          *entity.User
		wantErr       error
	}{
		{
			name:          "存在するユーザを正しく取得できる",
			spotifyUserID: "existing_user_spotify",
			want: &entity.User{
				ID:            "existing_user",
				SpotifyUserID: "existing_user_spotify",
				DisplayName:   "display_name",
				DeviceID:      "device_id",
			},
			wantErr: nil,
		},
		{
			name:          "存在しないspotifyUserIDの場合ErrUserNotFound",
			spotifyUserID: "not_found",
			want:          nil,
			wantErr:       entity.ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &UserRepository{dbMap: dbMap}
			got, err := r.FindBySpotifyUserID(tt.spotifyUserID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindBySpotifyUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("FindBySpotifyUserID() diff=%v", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestUserRepository_Store(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(userDTO{}, "users")
	truncateTable(t, dbMap)
	if err := dbMap.Insert(&userDTO{
		ID:            "existing_user",
		SpotifyUserID: "existing_user_spotify",
		DisplayName:   "display_name",
		DeviceID:      "device_id",
	}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		user    *entity.User
		wantErr error
	}{
		{
			name: "新規のユーザを正しく保存できる",
			user: &entity.User{
				ID:            "new_user",
				SpotifyUserID: "new_spotify_user",
				DisplayName:   "displayName",
				DeviceID:      "device_id",
			},
			wantErr: nil,
		},
		{
			name: "既に登録済みのユーザの場合ErrUserAlreadyExistedエラーを返す",
			user: &entity.User{
				ID:            "new_user",
				SpotifyUserID: "new_spotify_user",
				DisplayName:   "displayName",
				DeviceID:      "device_id",
			},
			wantErr: entity.ErrUserAlreadyExisted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &UserRepository{
				dbMap: dbMap,
			}
			if err := r.Store(tt.user); !errors.Is(err, tt.wantErr) {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
