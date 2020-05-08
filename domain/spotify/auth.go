//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package spotify

import (
	"golang.org/x/oauth2"
)

// Auth はSpotify OAuthに関連したAPIを呼び出すためのインターフェイスです。
type Auth interface {
	GetAuthURL(state string) string
	Exchange(code string) (*oauth2.Token, error)
	Refresh(token *oauth2.Token) (*oauth2.Token, error)
}
