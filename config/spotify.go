package config

import "os"

// Spotify はSpotifyに関連する設定を表します。
type Spotify struct {
	clientID     string
	clientSecret string
	redirectURL  string
}

// ClientID はSpotify Web APIのClient IDを取得します。
func (s Spotify) ClientID() string {
	return s.clientID
}

// ClientSecret はSpotify Web APIのClient Secretを取得します。
func (s Spotify) ClientSecret() string {
	return s.clientSecret
}

// RedirectURL はSpotify Web APIのRedirect URLを取得します。
func (s Spotify) RedirectURL() string {
	return s.redirectURL
}

// NewSpotify はSpotifyに関連する設定を環境変数から取得してSpotify構造体を返します、
func NewSpotify() *Spotify {
	return &Spotify{
		clientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		clientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		redirectURL:  os.Getenv("SPOTIFY_REDIRECT_URL"),
	}
}
