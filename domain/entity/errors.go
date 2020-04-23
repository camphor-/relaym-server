package entity

import "errors"

var (
	// ErrUserAlreadyExisted はユーザが既に存在しているときのエラーを表します。
	ErrUserAlreadyExisted = errors.New("user has already existed")
)
