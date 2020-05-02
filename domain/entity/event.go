package entity

// Event はクライアントに送信するイベントを表します。
type Event struct {
	Type string `json:"type"`
}
