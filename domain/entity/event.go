package entity

// Event はクライアントに送信するイベントを表します。
type Event struct {
	Type string `json:"type"`
}

var (
	// EventPlay はセッションの再生が開始された際に発されるイベントです。
	EventPlay = &Event{
		Type: "PLAY",
	}

	// EventNextTrack はセッションの曲の再生が (正常に) 次の曲に移った際に発されるイベントです。
	// キューの現在再生している曲の位置が含まれます。
	EventNextTrack = &Event{
		Type: "NEXTTRACK",
	}

	// EventInterrupt はSpotifyの本体アプリ側で操作されて、Relaym側との同期が取れなくなったタイミングで発されるイベントです。
	// セッションは一時停止状態になり、RESUMEを送ることで再開されます。
	EventInterrupt = &Event{
		Type: "INTERRUPT",
	}
)
