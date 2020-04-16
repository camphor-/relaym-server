package entity

// StateTemp はSpotifyの認可時に一時的に保存しておく必要がある情報を表します。
type StateTemp struct {
	State       string
	RedirectURL string
}
