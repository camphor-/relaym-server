// +build integration

package spotify

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/camphor-/relaym-server/config"
	"github.com/camphor-/relaym-server/domain/service"
	"golang.org/x/oauth2"
)

func TestClient_GetTracksFromURI(t *testing.T) {
	tests := []struct {
		name          string
		uris          []string
		wantTrackName string // idsで全て同じidを
		wantErr       bool
	}{
		{
			name: "50曲を超える(60曲)TrackURIsを指定しても全てのTrackが取得される",
			uris: []string{
				"spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF",
				"spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF",
				"spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF",
				"spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF",
				"spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF",
				"spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF", "spotify:track:5uQ0vKy2973Y9IUCd1wMEF",
			},
			wantTrackName: "新芽探して",
			wantErr:       false,
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
			token, err := c.Refresh(token)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			ctx = service.SetTokenToContext(ctx, token)

			tracks, err := c.GetTracksFromURI(ctx, tt.uris)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTracksFromURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(tracks) != len(tt.uris) {
				t.Errorf("GetTracksFromURI() gotDefferentLengthTracks: got %v, want %v", len(tracks), len(tt.uris))
				return
			}

			for _, tr := range tracks {
				if tr.Name != tt.wantTrackName {
					t.Errorf("GetTracksFromURI() gotDifferentTrack: got %v, want %v", tr.Name, tt.wantTrackName)
					return
				}
			}
		})
	}
}
