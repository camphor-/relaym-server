package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewSpotify(t *testing.T) {
	tests := []struct {
		name string
		want *Spotify
	}{
		{
			name: "正しく環境変数が読み込める",
			want: &Spotify{
				clientID:     "INPUT_YOUR_CLIENT_ID",
				clientSecret: "INPUT_YOUR_CLIENT_SECRET",
				redirectURL:  "http://localhost.local:8080/api/v3/callback",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := cmp.AllowUnexported(Spotify{})
			if got := NewSpotify(); !cmp.Equal(got, tt.want, opt) {
				t.Errorf("NewSpotify() diff=%s", cmp.Diff(tt.want, got, opt))
			}
		})
	}
}
