package entity

import "github.com/google/uuid"

type (
	// User はログインしているユーザを表します。
	User struct {
		ID            string // IDは外部のパッケージで書き換えられると困るのでprivateにする
		SpotifyUserID string
		DisplayName   string
	}

	// SpotifyUser はSpotify APIのユーザ情報を表します。
	SpotifyUser struct {
		SpotifyUserID string
		DisplayName   string
		Product       string
	}
)

// NewUser はUserのポインタを生成する関数です。
func NewUser(spotifyUserID, displayName string) *User {
	return &User{
		ID:            uuid.New().String(),
		SpotifyUserID: spotifyUserID,
		DisplayName:   displayName,
	}
}

// SpotifyURI はユーザを一位に識別するURLを返します。
func (u *User) SpotifyURI() string {
	return "spotify:user:" + u.SpotifyUserID
}
