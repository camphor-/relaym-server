package repository

import "github.com/camphor-/relaym-server/domain/entity"

// User はユーザの永続化を担当するリポジトリです。
type Auth interface {
	Store(state *entity.StateTemp) error
	FindStateByState(state string) (*entity.StateTemp, error)
	Delete(state string) error
}
