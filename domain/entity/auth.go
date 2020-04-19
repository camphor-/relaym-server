package entity

// AuthState はSpotifyの認可時に一時的に保存しておく必要がある情報を表します。
type AuthState struct {
	State       string
	RedirectURL string
}
