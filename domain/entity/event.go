package entity

// Event はクライアントに送信するイベントを表します。
type Event struct {
	Type string `json:"type"`
}

var (
	// EventAddTrack はセッションに曲が追加された際に発されるイベントです。
	EventAddTrack = &Event{
		Type: "ADDTRACK",
	}

	// EventPlay はセッションの再生が開始された際に発されるイベントです。
	EventPlay = &Event{
		Type: "PLAY",
	}

	// EventPause はセッションが一時停止された際に発されるイベントです。
	EventPause = &Event{
		Type: "PAUSE",
	}

	// EventNextTrack はセッションの曲の再生が (正常に) 次の曲に移った際に発されるイベントです。
	// キューの現在再生している曲の位置が含まれます。
	EventNextTrack = &Event{
		Type: "NEXTTRACK",
	}

	// EventStop は全ての曲の再生が終了した際に発されるイベントです。
	EventStop = &Event{
		Type: "STOP",
	}

	// EventInterrupt はSpotifyの本体アプリ側で操作されて、Relaym側との同期が取れなくなったタイミングで発されるイベントです。
	// セッションは一時停止状態になり、RESUMEを送ることで再開されます。
	EventInterrupt = &Event{
		Type: "INTERRUPT",
	}
)
