package spotify

import (
	"fmt"

	"github.com/camphor-/relaym-server/config"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

// Client はSpotifyのWeb APIをコールするクライアントです、
type Client struct {
	auth spotify.Authenticator
}

// NewClient はClientのポインタを生成する関数です。
func NewClient(cfg *config.Spotify) *Client {
	auth := spotify.NewAuthenticator(cfg.RedirectURL(), spotify.ScopeUserReadPrivate)
	auth.SetAuthInfo(cfg.ClientID(), cfg.ClientSecret())
	return &Client{auth: auth}
}

// GetAuthURL はSpotifyの認証画面のURLを取得します。
func (c *Client) GetAuthURL(state string) string {
	return c.auth.AuthURL(state)
}

// Exchange は Authorization codeを使ってOAuthのアクセストークンを取得します。
// ref : https://developer.spotify.com/documentation/general/guides/authorization-guide/
func (c *Client) Exchange(code string) (*oauth2.Token, error) {
	token, err := c.auth.Exchange(code)
	if err != nil {
		return nil, fmt.Errorf("excahnge code: %w", err)
	}
	return token, nil
}
