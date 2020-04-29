package event

// Pusher はイベントを対応したセッションIDに向けて送信するインターフェースです。
type Pusher interface {
	Push(pushMsg *PushMessage)
}

// PushMessage はPusherで送信するメッセージを表します。
type PushMessage struct {
	SessionID string
	Msg       interface{}
}
