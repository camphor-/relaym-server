package config

import (
	"testing"
)

func TestDSN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{
			name: "テスト時の環境変数を読み込んでDSNの構成できる",
			want: "app:Password@tcp(127.0.0.1:3306)/relaym?parseTime=true&collation=utf8mb4_bin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DSN(); got != tt.want {
				t.Errorf("DSN() = %v, want %v", got, tt.want)
			}
		})
	}
}
