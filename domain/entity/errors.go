package entity

import "errors"

var (
	// ErrUserNotFound はユーザが存在しないエラーを表します。
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExisted はユーザが既に存在しているエラーを表します。
	ErrUserAlreadyExisted = errors.New("user has already existed")

	// ErrTokenNotFound はSpotifyのアクセストークンが存在しないエラーを表します。
	ErrTokenNotFound = errors.New("token not found")

	// ErrActiveDeviceNotFound は再生できるアクティブなデバイスが存在しないエラーを表します。
	ErrActiveDeviceNotFound = errors.New("active device not found")
	// ErrNonPremium はユーザがプレミアム会員ではないエラーを表します。
	ErrNonPremium = errors.New("non-premium user")
)
