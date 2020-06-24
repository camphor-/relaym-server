package config

import "testing"

func TestPort(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "正しくポートを取得できる",
			want: "8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Port(); got != tt.want {
				t.Errorf("Port() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSAllowOrigin(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "正しく取得できる",
			want: "http://relaym.local:3000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CORSAllowOrigin(); got != tt.want {
				t.Errorf("CORSAllowOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFrontendURL(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "正しく取得できる",
			want: "http://relaym.local:3000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FrontendURL(); got != tt.want {
				t.Errorf("FrontendURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
