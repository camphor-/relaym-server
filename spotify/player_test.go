// +build integration

package spotify

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/camphor-/relaym-server/domain/service"
	"golang.org/x/oauth2"

	"github.com/camphor-/relaym-server/config"
)

func TestClient_CurrentlyPlaying(t *testing.T) {
	tests := []struct {
		name    string
		want    bool
		wantErr bool
	}{
		{
			name:    "再生中ではないときfalse",
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(config.NewSpotify())
			token := &oauth2.Token{
				AccessToken:  "",
				TokenType:    "Bearer",
				RefreshToken: os.Getenv("SPOTIFY_REFRESH_TOKEN_FOR_TEST"),
				Expiry:       time.Now(),
			}
			token, err := c.Refresh(context.Background(), token)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			ctx = service.SetTokenToContext(ctx, token)

			got, err := c.CurrentlyPlaying(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("CurrentlyPlaying() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Playing != tt.want {
				t.Errorf("CurrentlyPlaying() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Play(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "現在の曲が再生される",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(config.NewSpotify())
			token := &oauth2.Token{
				AccessToken:  "",
				TokenType:    "Bearer",
				RefreshToken: os.Getenv("SPOTIFY_REFRESH_TOKEN_FOR_TEST"),
				Expiry:       time.Now(),
			}
			token, err := c.Refresh(context.Background(), token)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			ctx = service.SetTokenToContext(ctx, token)

			if err := c.Play(ctx, ""); (err != nil) != tt.wantErr {
				t.Errorf("Play() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Pause(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "曲が一時停止される",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(config.NewSpotify())
			token := &oauth2.Token{
				AccessToken:  "",
				TokenType:    "Bearer",
				RefreshToken: os.Getenv("SPOTIFY_REFRESH_TOKEN_FOR_TEST"),
				Expiry:       time.Now(),
			}
			token, err := c.Refresh(context.Background(), token)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			ctx = service.SetTokenToContext(ctx, token)
			if err := c.Pause(ctx, ""); (err != nil) != tt.wantErr {
				t.Errorf("Pause() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Enqueue(t *testing.T) {

	tests := []struct {
		name    string
		trackID string
		wantErr bool
	}{
		{
			name:    "曲をqueueに追加できる",
			trackID: "spotify:track:49BRCNV7E94s7Q2FUhhT3w",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(config.NewSpotify())
			token := &oauth2.Token{
				AccessToken:  "",
				TokenType:    "Bearer",
				RefreshToken: os.Getenv("SPOTIFY_REFRESH_TOKEN_FOR_TEST"),
				Expiry:       time.Now(),
			}
			token, err := c.Refresh(context.Background(), token)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			ctx = service.SetTokenToContext(ctx, token)

			if err := c.Enqueue(ctx, tt.trackID, ""); (err != nil) != tt.wantErr {
				t.Errorf("Enqueue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_SetRepeatMode(t *testing.T) {
	tests := []struct {
		name    string
		on      bool
		wantErr bool
	}{
		{
			name:    "リピートモードをオフにできる",
			on:      false,
			wantErr: false,
		},
		{
			name:    "リピートモードをオンにできる",
			on:      true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(config.NewSpotify())
			token := &oauth2.Token{
				AccessToken:  "",
				TokenType:    "Bearer",
				RefreshToken: os.Getenv("SPOTIFY_REFRESH_TOKEN_FOR_TEST"),
				Expiry:       time.Now(),
			}
			token, err := c.Refresh(context.Background(), token)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			ctx = service.SetTokenToContext(ctx, token)
			if err := c.SetRepeatMode(ctx, tt.on, ""); (err != nil) != tt.wantErr {
				t.Errorf("SetRepeatMode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// テスト前にSpotify側で「次に再生される曲」を積んでください
func TestClient_DeleteAllTracksInQueueTracks(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
	}{
		{
			name:    "正常に動作する",
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(config.NewSpotify())
			token := &oauth2.Token{
				AccessToken:  "",
				TokenType:    "Bearer",
				RefreshToken: os.Getenv("SPOTIFY_REFRESH_TOKEN_FOR_TEST"),
				Expiry:       time.Now(),
			}
			token, err := c.Refresh(context.Background(), token)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			ctx = service.SetTokenToContext(ctx, token)
			if err := c.DeleteAllTracksInQueue(ctx, "", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF"); err.Error() != tt.wantErr {
				t.Errorf("DeleteAllTracksInQueue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
