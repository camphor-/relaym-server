package spotify

import (
	"context"
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/config"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
)

// Client はSpotifyのWeb APIをコールするクライアントです。
type Client struct {
	auth  *spotifyauth.Authenticator
	cache *cache.Cache
}

// NewClient はClientのポインタを生成する関数です。
func NewClient(cfg *config.Spotify) *Client {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(cfg.RedirectURL()),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserReadPlaybackState, spotifyauth.ScopeUserModifyPlaybackState),
		spotifyauth.WithClientID(cfg.ClientID()),
		spotifyauth.WithClientSecret(cfg.ClientSecret()))
	return &Client{auth: auth, cache: cache.New(10*time.Minute, 20*time.Minute)}
}

// GetAuthURL はSpotifyの認証画面のURLを取得します。
func (c *Client) GetAuthURL(state string) string {
	return c.auth.AuthURL(state)
}

// Exchange は Authorization codeを使ってOAuthのアクセストークンを取得します。
// ref : https://developer.spotify.com/documentation/general/guides/authorization-guide/
func (c *Client) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.auth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("excahnge code: %w", err)
	}
	return token, nil
}

// Refresh はリフレッシュトークンを使用して新しいアクセストークンを取得します。
func (c *Client) Refresh(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	cli := spotify.New(c.auth.Client(ctx, token))
	newToken, err := cli.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh: %w", err)
	}
	return newToken, nil
}
