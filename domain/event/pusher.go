package event

type Pusher interface {
	Push(pushMsg *PushMessage)
}

type PushMessage struct {
	SessionID string
	Msg       interface{}
}
