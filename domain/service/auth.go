package service

import (
	"context"

	"golang.org/x/oauth2"
)

// ContextKey はContextに情報を保存するときのキーです。
type ContextKey string

var (
	userIDKey    ContextKey = "userIDKey"
	creatorIDKey ContextKey = "creatorIDKey"
	tokenKey     ContextKey = "tokenKey"
)

// SetUserIDToContext はユーザIDをContextにセットします。
func SetUserIDToContext(ctx context.Context, userID string) context.Context {
	if userID != "" {
		return context.WithValue(ctx, userIDKey, userID)
	}
	return ctx
}

// SetCreatorIDToContext はセッション作成者のIDをContextにセットします。
func SetCreatorIDToContext(ctx context.Context, userID string) context.Context {
	if userID != "" {
		return context.WithValue(ctx, creatorIDKey, userID)
	}
	return ctx
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

// GetCreatorIDFromContext はContextからセッション作成者のIDを取得します。
func GetCreatorIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(creatorIDKey)
	userID, ok := v.(string)
	return userID, ok
}

// GetTokenFromContext はContextからトークンを取得します。
func GetTokenFromContext(ctx context.Context) (*oauth2.Token, bool) {
	v := ctx.Value(tokenKey)
	token, ok := v.(*oauth2.Token)
	return token, ok
}

// NewContextFromContext は既存のContextに含まれるトークンなどをコピーした上で、新しいContextを生成します。
// これは、goroutine内のループなど、HTTPリクエスト終了後も生き残って欲しいContextを作るのに使われます。
func NewBackgroundContextFromContext(prevCtx context.Context) context.Context {
	ctx := context.Background()
	userId, ok := GetUserIDFromContext(prevCtx)
	if ok {
		ctx = SetUserIDToContext(ctx, userId)
	}
	creatorId, ok := GetCreatorIDFromContext(prevCtx)
	if ok {
		ctx = SetCreatorIDToContext(ctx, creatorId)
	}
	token, ok := GetTokenFromContext(prevCtx)
	if ok {
		ctx = SetTokenToContext(ctx, token)
	}
	return ctx

}
