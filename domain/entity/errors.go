package entity

import "errors"

var (
	// ErrUserNotFound はユーザが存在しないエラーを表します。
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExisted はユーザが既に存在しているエラーを表します。
	ErrUserAlreadyExisted = errors.New("user has already existed")

	// ErrSessionNotFound はセッションが存在しないエラーを表します。
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionAlreadyExisted はセッションが既に存在しているときのエラーを表します。
	ErrSessionAlreadyExisted = errors.New("session has already existed")
	// ErrSessionAllTracksFinished はセッションに追加された全てのトラックの再生が全て終了しているエラーを表します。
	ErrSessionAllTracksFinished = errors.New("all tracks has already finished")
	// ErrSessionPlayingDifferentTrack はキュー先頭の曲と異なる曲が再生されているエラーを表します。
	ErrSessionPlayingDifferentTrack = errors.New("session is playing different track from queue")

	// ErrUserIsNotSessionCreator はユーザがセッションの作成者でないときのエラーを表します。
	ErrUserIsNotSessionCreator = errors.New("user is not session's creator")

	// ErrQueueTrackNotFound はセッションに紐付くQueueTrackが存在しないエラーを表します。
	ErrQueueTrackNotFound = errors.New("queue track not found")

	// ErrNextQueueTrackNotFound は次に再生すべきQueueTrackが存在しないエラーを表します。
	ErrNextQueueTrackNotFound = errors.New("next queue track not found")

	// ErrTokenNotFound はSpotifyのアクセストークンが存在しないエラーを表します。
	ErrTokenNotFound = errors.New("token not found")

	// ErrActiveDeviceNotFound は再生できるアクティブなデバイスが存在しないエラーを表します。
	ErrActiveDeviceNotFound = errors.New("active device not found")
	// ErrNonPremium はユーザがプレミアム会員ではないエラーを表します。
	ErrNonPremium = errors.New("non-premium user")

	// ErrInvalidStateType は不正なstate typeであるというエラーを表します。
	ErrInvalidStateType = errors.New("invalid state type")

	// ErrChangeSessionStateNotPermit はセッションのステートの状態遷移が許可されていない場合のエラーを表します。
	ErrChangeSessionStateNotPermit = errors.New("change session state is not permits")

	// ErrLoginSessionNotFound はセッション(login)が存在しないエラーを表します。
	ErrLoginSessionNotFound = errors.New("loginSession not found")
	// ErrLoginSessionAlreadyExisted はセッション(login)が既に存在しているときのエラーを表します。
	ErrLoginSessionAlreadyExisted = errors.New("loginSession has already existed")
)
