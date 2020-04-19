package repository

import "github.com/camphor-/relaym-server/domain/entity"

// Auth は認証・認可に関するの永続化を担当するリポジトリです。
type Auth interface {
	Store(authState *entity.AuthState) error
	FindStateByState(state string) (*entity.AuthState, error)
	Delete(state string) error
}
