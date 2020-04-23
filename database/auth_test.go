package database

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"

	_ "github.com/go-sql-driver/mysql"

	"golang.org/x/oauth2"

	"github.com/camphor-/relaym-server/domain/entity"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestAuthRepository_StoreORUpdateToken(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(spotifyAuthDTO{}, "spotify_auth")
	truncateTable(t, dbMap)
	if err := dbMap.Insert(&spotifyAuthDTO{
		SpotifyUserID: "existing_user",
		AccessToken:   "existing_access_token",
		RefreshToken:  "existing_refresh_token",
		Expiry:        time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		spotifyUserID string
		token         *oauth2.Token
		wantErr       bool
	}{
		{
			name:          "新規ユーザのトークンを保存できる",
			spotifyUserID: "new_user",
			token: &oauth2.Token{
				AccessToken:  "new_user_access_token",
				TokenType:    "Bearer",
				RefreshToken: "new_user_refresh_token",
				Expiry:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},

		{
			name:          "既存ユーザのトークンを更新できる",
			spotifyUserID: "existing_user",
			token: &oauth2.Token{
				AccessToken:  "update_user_access_token",
				TokenType:    "Bearer",
				RefreshToken: "update_user_refresh_token",
				Expiry:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := AuthRepository{dbMap: dbMap}
			if err := r.StoreORUpdateToken(tt.spotifyUserID, tt.token); (err != nil) != tt.wantErr {
				t.Errorf("StoreORUpdateToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got, err := r.GetTokenBySpotifyUserID(tt.spotifyUserID)
				if err != nil {
					t.Fatal(err)
				}
				opt := cmpopts.IgnoreUnexported(oauth2.Token{})
				if !cmp.Equal(got, tt.token, opt) {
					t.Errorf("StoreState() diff=%v", cmp.Diff(tt.token, got, opt))
				}
			}
		})
	}
}

func TestAuthRepository_GetTokenBySpotifyUserID(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(spotifyAuthDTO{}, "spotify_auth")
	truncateTable(t, dbMap)
	if err := dbMap.Insert(&spotifyAuthDTO{
		SpotifyUserID: "get_user",
		AccessToken:   "get_access_token",
		RefreshToken:  "get_refresh_token",
		Expiry:        time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		spotifyUserID string
		want          *oauth2.Token
		wantErr       bool
	}{
		{
			name:          "保存してあるトークンを取得できる",
			spotifyUserID: "get_user",
			want: &oauth2.Token{
				AccessToken:  "get_access_token",
				RefreshToken: "get_refresh_token",
				TokenType:    "Bearer",
				Expiry:       time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:          "存在しないユーザのトークンを取得しようとするとエラーになる",
			spotifyUserID: "not_found_user",
			want:          nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := AuthRepository{dbMap: dbMap}
			got, err := r.GetTokenBySpotifyUserID(tt.spotifyUserID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTokenBySpotifyUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			opt := cmpopts.IgnoreUnexported(oauth2.Token{})
			if !cmp.Equal(got, tt.want, opt) {
				t.Errorf("GetTokenBySpotifyUserID() diff=%v", cmp.Diff(tt.want, got, opt))
			}
		})
	}
}

func TestAuthRepository_Store(t *testing.T) {
	tests := []struct {
		name    string
		state   *entity.AuthState
		wantErr bool
	}{
		{
			name: "正しく保存できる",
			state: &entity.AuthState{
				State:       uuid.New().String(),
				RedirectURL: "https://example.com",
			},
			wantErr: false,
		},
	}
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := NewAuthRepository(dbMap)
			if err := r.StoreState(tt.state); (err != nil) != tt.wantErr {
				t.Errorf("StoreState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := r.FindStateByState(tt.state.State)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, tt.state) {
					t.Errorf("StoreState() got = %v, want %v", got, tt.state)
				}
			}
		})
	}
}

func TestAuthRepository_FindStateByState(t *testing.T) {
	tests := []struct {
		name    string
		state   string
		want    *entity.AuthState
		wantErr bool
	}{
		{
			name:  "存在するStateを正しく取得できる",
			state: "state",
			want: &entity.AuthState{
				State:       "state",
				RedirectURL: "https://example.com",
			},
			wantErr: false,
		},
		{
			name:    "存在しないStateを取得しようとするとエラーになる",
			state:   "not_found",
			want:    nil,
			wantErr: true,
		},
	}
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	r := NewAuthRepository(dbMap)

	// Prepare
	truncateTable(t, dbMap)
	if err := r.StoreState(&entity.AuthState{
		State:       "state",
		RedirectURL: "https://example.com",
	}); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.FindStateByState(tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindStateByState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("FindStateByState() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthRepository_Delete(t *testing.T) {

	tests := []struct {
		name    string
		state   string
		wantErr bool
	}{
		{
			name:    "存在するStateを正しく削除できる",
			state:   "state",
			wantErr: false,
		},
		{
			name:    "存在しないStateを削除してもエラーにならない",
			state:   "not_found",
			wantErr: false,
		},
	}

	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	r := NewAuthRepository(dbMap)

	// Prepare
	truncateTable(t, dbMap)
	if err := r.StoreState(&entity.AuthState{
		State:       "state",
		RedirectURL: "https://example.com",
	}); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.Delete(tt.state); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
