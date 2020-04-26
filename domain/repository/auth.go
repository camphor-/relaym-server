//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package repository

import (
	"github.com/camphor-/relaym-server/domain/entity"
	"golang.org/x/oauth2"
)

// Auth は認証・認可に関する永続化を担当するリポジトリです。
type Auth interface {
	StoreORUpdateToken(userID string, token *oauth2.Token) error
	GetTokenByUserID(userID string) (*oauth2.Token, error)
	StoreSession(sessionID, userID string) error
	GetUserIDFromSession(sessionID string) (string, error)

	StoreState(authState *entity.AuthState) error
	FindStateByState(state string) (*entity.AuthState, error)
	DeleteState(state string) error
}
