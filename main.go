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
	"github.com/camphor-/relaym-server/web/ws"

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

	hub := ws.NewHub()
	go hub.Run()

	spotifyCFG := config.NewSpotify()
	spotifyCli := spotify.NewClient(spotifyCFG)

	authRepo := database.NewAuthRepository(dbMap)
	userRepo := database.NewUserRepository(dbMap)
	sessionRepo := database.NewSessionRepository(dbMap)
	userUC := usecase.NewUserUseCase(spotifyCli, userRepo)
	authUC := usecase.NewAuthUseCase(spotifyCli, spotifyCli, authRepo, userRepo)
	sessionUC := usecase.NewSessionUseCase(sessionRepo, userRepo, spotifyCli, hub)
	trackUC := usecase.NewTrackUseCase(spotifyCli)

	s := web.NewServer(authUC, userUC, sessionUC, trackUC, hub)

	// シグナルを受け取れるようにgoroutine内でサーバを起動する
	go func() {
		if err := s.Start(":" + config.Port()); err != nil {
			s.Logger.Infof("shutting down the server with error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	s.Logger.Infof("SIGNAL %d received, then shutting down...\n", <-quit)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		s.Logger.Fatal(err)
	}
}
