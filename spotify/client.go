package spotify

import (
	"fmt"

	"github.com/camphor-/relaym-server/config"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

type Client struct {
	cli spotify.Client
}

func NewClient(cfg *config.Spotify, token *oauth2.Token) *Client {
	auth := spotify.NewAuthenticator(cfg.RedirectURL(), spotify.ScopeUserReadPrivate)
	cli := auth.NewClient(token)
	return &Client{cli: cli}
}

// Authenticater はSpotifyのWeb APIをコールするクライアントです、
type Authenticater struct {
	auth spotify.Authenticator
}

// NewAuthenticater はAuthenticaterのポインタを生成する関数です。
func NewAuthenticater(cfg *config.Spotify) *Authenticater {
	auth := spotify.NewAuthenticator(cfg.RedirectURL(), spotify.ScopeUserReadPrivate)
	auth.SetAuthInfo(cfg.ClientID(), cfg.ClientSecret())
	return &Authenticater{auth: auth}
}

// GetAuthURL はSpotifyの認証画面のURLを取得します。
func (c *Authenticater) GetAuthURL(state string) string {
	return c.auth.AuthURL(state)
}

// Exchange は Authorization codeを使ってOAuthのアクセストークンを取得します。
// ref : https://developer.spotify.com/documentation/general/guides/authorization-guide/
func (c *Authenticater) Exchange(code string) (*oauth2.Token, error) {
	token, err := c.auth.Exchange(code)
	if err != nil {
		return nil, fmt.Errorf("excahnge code: %w", err)
	}
	return token, nil
}
