package usecase

import (
	"fmt"

	"github.com/camphor-/relaym-server/domain/spotify"
	"github.com/google/uuid"
)

type AuthUseCase struct {
	authCli spotify.Auth
}

func NewAuthUseCase(authCli spotify.Auth) *AuthUseCase {
	return &AuthUseCase{authCli: authCli}
}

func (u *AuthUseCase) GetAuthURL(redirectURL string) string {
	state := uuid.New().String()
	// TODO stateとredirectURLを保存しておく
	return u.authCli.GetAuthURL(state)
}

// Authorization はcodeを使って認可をチェックします。
// 認可に成功した場合はフロントエンドのリダイレクトURLを返します。
func (u *AuthUseCase) Authorization(state, code string) (string, error) {
	// storedState := "hoge" // TODO どっかにGetAuthURL()で保存したstateとredirectを持ってくる
	// if storedState != state {
	// 	return "", fmt.Errorf("redirect state param doesn't match storedState=%s, state=%s", storedState, state)
	// }

	token, err := u.authCli.Exchange(code)
	if err != nil {
		return "", fmt.Errorf("exchange and get oauth2 token: %w", err)
	}
	// TODO : 手に入れたアクセストークンをDBに保存する
	fmt.Printf("%#v\n", token)

	// TODO リダイレクトURLを返す
	return "", nil
}
