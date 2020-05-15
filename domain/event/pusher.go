//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE

package event

import "github.com/camphor-/relaym-server/domain/entity"

// Pusher はイベントを対応したセッションIDに向けて送信するインターフェースです。
type Pusher interface {
	Push(pushMsg *PushMessage)
}

// PushMessage はPusherで送信するメッセージを表します。
type PushMessage struct {
	SessionID string
	Msg       *entity.Event
}
