package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/oauth2"
)

func TestGetUserIDFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		ctx   context.Context
		want  string
		want1 bool
	}{
		{
			name:  "userIDがセットされているとき正しくuserIDを取得できる",
			ctx:   SetUserIDToContext(context.Background(), "userID"),
			want:  "userID",
			want1: true,
		},
		{
			name:  "userIDがセットされていないとfalseが帰る",
			ctx:   context.Background(),
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetUserIDFromContext(tt.ctx)
			if got != tt.want {
				t.Errorf("GetUserIDFromContext() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetUserIDFromContext() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetTokenFromContext(t *testing.T) {
	t.Parallel()

	token := &oauth2.Token{
		AccessToken:  "access_token",
		TokenType:    "Bearer",
		RefreshToken: "refresh_token",
		Expiry:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name  string
		ctx   context.Context
		want  *oauth2.Token
		want1 bool
	}{
		{
			name:  "tokenがセットされているとき正しくtokenを取得できる",
			ctx:   SetTokenToContext(context.Background(), token),
			want:  token,
			want1: true,
		},
		{
			name:  "tokenがセットされていないとfalseが帰る",
			ctx:   context.Background(),
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetTokenFromContext(tt.ctx)

			opt := cmpopts.IgnoreUnexported(oauth2.Token{})
			if !cmp.Equal(got, tt.want, opt) {
				t.Errorf("GetTokenFromContext() diff=%s", cmp.Diff(tt.want, got, opt))
			}
			if got1 != tt.want1 {
				t.Errorf("GetTokenFromContext() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
