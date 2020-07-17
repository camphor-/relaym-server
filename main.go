package main

import (
	"context"

	"os"
	"os/signal"
	"time"

	"github.com/camphor-/relaym-server/config"
	"github.com/camphor-/relaym-server/database"
	"github.com/camphor-/relaym-server/log"
	"github.com/camphor-/relaym-server/spotify"
	"github.com/camphor-/relaym-server/usecase"
	"github.com/camphor-/relaym-server/web"
	"github.com/camphor-/relaym-server/web/ws"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	logger := log.New()

	dbMap, err := database.NewDB()
	if err != nil {
		logger.Fatal(err)
	}
	defer func() {
		err := dbMap.Db.Close()
		if err != nil {
			logger.Fatal(err)
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
	authUC := usecase.NewAuthUseCase(spotifyCli, spotifyCli, authRepo, userRepo, sessionRepo)
	sessionTimerUC := usecase.NewSessionTimerUseCase(sessionRepo, spotifyCli, hub)
	sessionUC := usecase.NewSessionUseCase(sessionRepo, userRepo, spotifyCli, spotifyCli, spotifyCli, hub, sessionTimerUC)
	sessionStateUC := usecase.NewSessionStateUseCase(sessionRepo, spotifyCli, hub, sessionTimerUC)
	trackUC := usecase.NewTrackUseCase(spotifyCli)
	batchUC := usecase.NewBatchUseCase(sessionRepo, hub)

	s := web.NewServer(authUC, userUC, sessionUC, sessionStateUC, trackUC, batchUC, hub)

	// シグナルを受け取れるようにgoroutine内でサーバを起動する
	go func() {
		if err := s.Start(":" + config.Port()); err != nil {
			logger.Infof("shutting down the server with error: %v", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	logger.Infof("SIGNAL %d received, then shutting down...", <-quit)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		logger.Fatal(err)
	}
}
