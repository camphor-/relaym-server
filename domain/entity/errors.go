package entity

import "errors"

var (
	// ErrUserNotFound はユーザが存在しないエラーを表します。
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExisted はユーザが既に存在しているときのエラーを表します。
	ErrUserAlreadyExisted = errors.New("user has already existed")
)
