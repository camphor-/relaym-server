package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/camphor-/relaym-server/config"
	"github.com/camphor-/relaym-server/database"
	"github.com/camphor-/relaym-server/spotify"
	"github.com/camphor-/relaym-server/usecase"
	"github.com/camphor-/relaym-server/web"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dbMap, err := database.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := dbMap.Db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	spotifyCFG := config.NewSpotify()
	spotifyCli := spotify.NewClient(spotifyCFG)

	authRepo := database.NewAuthRepository(dbMap)
	userRepo := database.NewUserRepository(dbMap)
	userUC := usecase.NewUserUseCase(userRepo)
	authUC := usecase.NewAuthUseCase(spotifyCli, spotifyCli, authRepo, userRepo)
	trackUC := usecase.NewTrackUseCase(spotifyCli)

	s := web.NewServer(authUC, userUC, trackUC)

	// シグナルを受け取れるようにgoroutine内でサーバを起動する
	go func() {
		if err := s.Start(":8080"); err != nil { // TODO : ポート番号を環境変数から読み込む
			s.Logger.Infof("shutting down the server with error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	s.Logger.Infof("SIGNAL %d received, then shutting down...\n", <-quit)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		s.Logger.Fatal(err)
	}
}
