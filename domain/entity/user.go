package entity

type (
	// User はログインしているユーザを表します
	User struct {
		id          string
		spotifyUser *SpotifyUser
	}

	SpotifyUser struct {
		id          string
		displayName string
		product     string
	}
)

// IsPremium はユーザがプレミアム会員かどうかチェックします。
func (u *User) IsPremium() bool {
	if u.spotifyUser == nil {
		return false
	}
	return u.spotifyUser.product == "premium"
}

// SpotifyURI はユーザを一位に識別するURLを返します。
func (u *User) SpotifyURI() string {
	if u.spotifyUser == nil {
		return ""
	}
	return "spotify:user:" + u.spotifyUser.id
}
