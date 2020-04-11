package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/camphor-/relaym-server/web"
)

func main() {
	e := web.NewRouter()

	// シグナルを受け取れるようにgoroutine内でサーバを起動する
	go func() {
		if err := e.Start(":8080"); err != nil { // TODO : ポート番号を環境変数から読み込む
			e.Logger.Infof("shutting down the server with error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	e.Logger.Infof("SIGNAL %d received, then shutting down...\n", <-quit)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
