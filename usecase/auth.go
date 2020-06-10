package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/camphor-/relaym-server/domain/spotify"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// AuthUseCase は認証・認可に関するユースケースです。
type AuthUseCase struct {
	authCli     spotify.Auth
	userCli     spotify.User
	repo        repository.Auth
	userRepo    repository.User
	sessionRepo repository.Session
}

// NewAuthUseCase はAuthUseCaseのポインタを生成します。
func NewAuthUseCase(authCli spotify.Auth, userCli spotify.User, repo repository.Auth, userRepo repository.User, sessionRepo repository.Session) *AuthUseCase {
	return &AuthUseCase{authCli: authCli, userCli: userCli, repo: repo, userRepo: userRepo, sessionRepo: sessionRepo}
}

// GetAuthURL はSpotifyの認可画面のリンクを生成します。
// CSRF対策のためにstateを保存しておいて、callbackを受け取った時に正当性を確認する必要がある。
func (u *AuthUseCase) GetAuthURL(redirectURL string) (string, error) {
	state := uuid.New().String()
	st := &entity.AuthState{
		State:       state,
		RedirectURL: redirectURL,
	}
	if err := u.repo.StoreState(st); err != nil {
		return "", fmt.Errorf("store state for authorization: %w", err)
	}
	return u.authCli.GetAuthURL(state), nil
}

// Authorization はcodeを使って認可をチェックします。
// 認可に成功した場合はフロントエンドのリダイレクトURLとセッションIDを返します。
func (u *AuthUseCase) Authorization(state, code string) (string, string, error) {
	storedState, err := u.repo.FindStateByState(state)
	if err != nil {
		return "", "", fmt.Errorf("find temp state state=%s: %w", state, err)
	}

	token, err := u.authCli.Exchange(code)
	if err != nil {
		return "", "", fmt.Errorf("exchange and get oauth2 token: %w", err)
	}

	ctx := service.SetTokenToContext(context.Background(), token)
	userID, err := u.createUserIfNotExists(ctx)
	if err != nil {
		return "", "", fmt.Errorf("get or create user: %w", err)
	}

	if err := u.repo.StoreORUpdateToken(userID, token); err != nil {
		return "", "", fmt.Errorf("store or update oauth token though repo userID=%s: %w", userID, err)
	}

	sessionID := uuid.New().String()
	if err := u.repo.StoreSession(sessionID, userID); err != nil {
		return "", "", fmt.Errorf("store session sessionID=%s userID=%s : %w", sessionID, userID, err)
	}

	// Stateを削除するのが失敗してもログインは成功しているので、エラーを返さない
	if err := u.repo.DeleteState(state); err != nil {
		log.Printf("Failed to delete state state=%s: %v\n", state, err)
		return storedState.RedirectURL, sessionID, nil
	}

	return storedState.RedirectURL, sessionID, nil
}

// GetTokenByUserID は対応したユーザのアクセストークンを取得します。
func (u *AuthUseCase) GetTokenByUserID(userID string) (*oauth2.Token, error) {
	token, err := u.repo.GetTokenByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("get oauth token userID=%s: %w", userID, err)
	}
	return token, nil
}

// createUserIfNotExists はユーザが存在していなかったら新規に作成しIDを返します。
func (u *AuthUseCase) createUserIfNotExists(ctx context.Context) (string, error) {
	spotifyUser, err := u.userCli.GetMe(ctx)
	if err != nil {
		return "", fmt.Errorf("get my info from Spotify: %w", err)
	}

	spotifyUserID := spotifyUser.SpotifyUserID

	user := entity.NewUser(spotifyUserID, spotifyUser.DisplayName)

	if err := u.userRepo.Store(user); err != nil {
		if errors.Is(err, entity.ErrUserAlreadyExisted) {
			existing, err := u.userRepo.FindBySpotifyUserID(spotifyUserID)
			if err != nil {
				return "", fmt.Errorf("find already existing user spotifyUserID=%s: %w", spotifyUserID, err)
			}
			return existing.ID, nil
		}
		return "", fmt.Errorf("store user through repo userID=%s: %w", user.ID, err)
	}

	return user.ID, nil
}

// GetUserIDFromSession はセッションIDから対応するユーザIDを返します。
func (u *AuthUseCase) GetUserIDFromSession(sessionID string) (string, error) {
	userID, err := u.repo.GetUserIDFromSession(sessionID)
	if err != nil {
		return "", fmt.Errorf("get user from session sessionID=%s: %w", sessionID, err)
	}
	return userID, nil
}

// RefreshAccessToken はリフレッシュトークンを使用してアクセストークンを更新し保存します。
func (u *AuthUseCase) RefreshAccessToken(userID string, token *oauth2.Token) (*oauth2.Token, error) {
	if token.Valid() {
		return token, nil
	}
	newToken, err := u.authCli.Refresh(token)
	if err != nil {
		return nil, fmt.Errorf("refresh access token through spotify client: %w", err)
	}

	if err := u.repo.StoreORUpdateToken(userID, newToken); err != nil {
		return nil, fmt.Errorf("update new token: %w", err)
	}
	return newToken, nil
}

// GetTokenBySessionID は指定されたidからsessionの持つcreatorのtokenを返します
func (u *AuthUseCase) GetTokenBySessionID(sessionID string) (*oauth2.Token, string, error) {
	token, creatorID, err := u.sessionRepo.FindCreatorTokenBySessionID(sessionID)
	if err != nil {
		return nil, "", fmt.Errorf("FindCreatorTokenBySessionID: sessionID=%s: %w", sessionID, err)
	}

	return token, creatorID, nil
}
