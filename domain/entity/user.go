package entity

type (
	// User はログインしているユーザを表します。
	User struct {
		id          string // IDは外部のパッケージで書き換えられると困るのでprivateにする
		SpotifyUser SpotifyUser
	}

	SpotifyUser struct {
		id          string
		product     string
		DisplayName string
	}
)

// NewUser はUserのポインタを生成する関数です。
func NewUser(id string) *User {
	return &User{id: id}
}

// ID はユーザIDを返します。
func (u *User) ID() string {
	return u.id
}

// IsPremium はユーザがプレミアム会員かどうかチェックします。
func (u *User) IsPremium() bool {
	return u.SpotifyUser.product == "premium"
}

// SpotifyURI はユーザを一位に識別するURLを返します。
func (u *User) SpotifyURI() string {
	return "spotify:user:" + u.SpotifyUser.id
}
