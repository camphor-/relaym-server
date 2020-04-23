package usecase

import (
	"fmt"
	"log"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/camphor-/relaym-server/domain/spotify"

	"github.com/google/uuid"
)

// AuthUseCase は認証・認可に関するユースケースです。
type AuthUseCase struct {
	authCli spotify.Auth
	repo    repository.Auth
}

// NewAuthUseCase はAuthUseCaseのポインタを生成します。
func NewAuthUseCase(authCli spotify.Auth, repo repository.Auth) *AuthUseCase {
	return &AuthUseCase{authCli: authCli, repo: repo}
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
// 認可に成功した場合はフロントエンドのリダイレクトURLを返します。
func (u *AuthUseCase) Authorization(state, code string) (string, error) {
	storedState, err := u.repo.FindStateByState(state)
	if err != nil {
		return "", fmt.Errorf("find temp state state=%s: %w", state, err)
	}

	token, err := u.authCli.Exchange(code)
	if err != nil {
		return "", fmt.Errorf("exchange and get oauth2 token: %w", err)
	}
	// TODO : SpotifyUserIDを取得する
	spotifyUserID := "spotifyUserID"
	if err := u.repo.StoreORUpdateToken(spotifyUserID, token); err != nil {
		return "", fmt.Errorf("store or update oauth token though repo spotifyUserID=%s: %w", spotifyUserID, err)
	}
	fmt.Printf("%#v\n", token)

	// Stateを削除するのが失敗してもログインは成功しているので、エラーを返さない
	if err := u.repo.Delete(state); err != nil {
		log.Printf("Failed to delete state state=%s: %v\n", state, err)
		return storedState.RedirectURL, nil
	}

	return storedState.RedirectURL, nil
}
