package repository

import "github.com/camphor-/relaym-server/domain/entity"

// User はユーザの永続化を担当するリポジトリです。
type User interface {
	FindByID(id string) (*entity.User, error)
}
