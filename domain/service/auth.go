package service

import (
	"context"

	"golang.org/x/oauth2"
)

// ContextKey はContextに情報を保存するときのキーです。
type ContextKey string

var (
	userIDKey ContextKey = "userIDKey"
	tokenKey  ContextKey = "tokenKey"
)

// SetUserIDToContext はユーザIDをContextにセットします。
func SetUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// SetTokenToContext はトークンをContextにセットします。
func SetTokenToContext(ctx context.Context, token *oauth2.Token) context.Context {
	return context.WithValue(ctx, tokenKey, token)

}

// GetUserIDFromContext はContextからユーザIDを取得します。
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	userID, ok := v.(string)
	return userID, ok
}

// GetTokenFromContext はContextからトークンを取得します。
func GetTokenFromContext(ctx context.Context) (*oauth2.Token, bool) {
	v := ctx.Value(tokenKey)
	token, ok := v.(*oauth2.Token)
	return token, ok
}