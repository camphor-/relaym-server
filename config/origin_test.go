package config

import (
	"os"
	"testing"
)

func Test_origin_IsReliableOrigin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		origin string
		env    string
		want   bool
	}{
		{
			name:   "Cloudflare Pages: 正しいデプロイプレビューのURLならtrue",
			origin: "https://5c927597.relaym.pages.dev",
			env:    "dev",
			want:   true,
		},
		{
			name:   "Cloudflare Pages: 別ののURLならfalse",
			origin: "https://5c927597.relaym2.pages.dev",
			env:    "dev",
			want:   false,
		},
		{
			name:   "Cloudflare Pages: 前後に変な文字が入ってもfalse",
			origin: "https://evil.example.com/https://5c927597.relaym2.pages.dev/hoge",
			env:    "dev",
			want:   false,
		},
		{
			name:   "ローカル環境は何でも良い",
			origin: "http://relaym.local:3000",
			env:    "local",
			want:   true,
		},
		{
			name:   "本番環境では、FRONTEND_URLと同じ値のときはtrueになる",
			origin: "http://relaym.local:3000", // test時にセットされるFRONTEND_URLの環境変数がこれになっている。本番環境では本番のURLが環境変数で渡される。
			env:    "prod",
			want:   true,
		},
		{
			name:   "本番環境では、FRONTEND_URLと違う値のときはfalseになる",
			origin: "http://relaym.local:3001",
			env:    "prod",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalEnv := os.Getenv("ENV")
			os.Setenv("ENV", tt.env)
			defer func() {
				os.Setenv("ENV", originalEnv)
			}()
			if got := IsReliableOrigin(tt.origin); got != tt.want {
				t.Errorf("IsReliableOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
