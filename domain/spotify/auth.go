package spotify

import "golang.org/x/oauth2"

type Auth interface {
	GetAuthURL(state string) string
	Exchange(code string) (*oauth2.Token, error)
}
