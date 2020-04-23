package database

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/go-gorp/gorp/v3"
)

var _ repository.Auth = &AuthRepository{}

var sessionStore = map[string]string{}

// AuthRepository は repository.AuthRepository を満たす構造体です
type AuthRepository struct {
	dbMap *gorp.DbMap
}

// NewAuthRepository はAuthRepositoryのポインタを生成する関数です
func NewAuthRepository(dbMap *gorp.DbMap) *AuthRepository {
	dbMap.AddTableWithName(stateDTO{}, "auth_states")
	dbMap.AddTableWithName(spotifyAuthDTO{}, "spotify_auth")
	return &AuthRepository{dbMap: dbMap}
}

// StoreORUpdateToken は既にトークンが存在する場合は更新し、存在しない場合は新規に保存します。
func (r AuthRepository) StoreORUpdateToken(spotifyUserID string, token *oauth2.Token) error {
	query := `INSERT INTO spotify_auth (user_id, access_token, refresh_token, expiry)
				VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
				access_token = VALUES(access_token), refresh_token = VALUES(refresh_token), expiry = VALUES(expiry)`
	if _, err := r.dbMap.Exec(query, spotifyUserID, token.AccessToken, token.RefreshToken, token.Expiry); err != nil {
		return fmt.Errorf("insert to spotify_auth table: %w", err)
	}
	return nil
}

// GetTokenByUserID は与えられたユーザのOAuth2のトークンを取得します。
func (r AuthRepository) GetTokenByUserID(userID string) (*oauth2.Token, error) {
	var dto spotifyAuthDTO
	query := "SELECT user_id, access_token, refresh_token, expiry from spotify_auth WHERE user_id=?"
	if err := r.dbMap.SelectOne(&dto, query, userID); err != nil {
		return nil, fmt.Errorf("select spotify auth: %w", err)
	}
	return &oauth2.Token{
		AccessToken:  dto.AccessToken,
		TokenType:    "Bearer",
		RefreshToken: dto.RefreshToken,
		Expiry:       dto.Expiry,
	}, nil
}

// StoreSession はセッション情報を保存します。
// TODO セッションの保存をMySQLにするかRedisにするか決めてないので一旦インメモリで持つ。
func (r AuthRepository) StoreSession(sessionID, userID string) error {
	sessionStore[sessionID] = userID
	return nil
}

// GetSession はセッションIDからユーザIDを取得します。
func (r AuthRepository) GetUserIDFromSession(sessionID string) (string, error) {
	if userID, ok := sessionStore[sessionID]; ok {
		return userID, nil
	}
	return "", errors.New("session not found")
}

// StoreState はauthStateを保存します。
func (r AuthRepository) StoreState(authState *entity.AuthState) error {
	dto := &stateDTO{
		State:       authState.State,
		RedirectURL: authState.RedirectURL,
	}
	if err := r.dbMap.Insert(dto); err != nil {
		return fmt.Errorf("insert auth authState: %w", err)
	}
	return nil
}

// FindStateByState はstateをキーしてStateTempを取得する。
func (r AuthRepository) FindStateByState(state string) (*entity.AuthState, error) {
	var dto stateDTO
	if err := r.dbMap.SelectOne(&dto, "SELECT state, redirect_url from auth_states WHERE state=?", state); err != nil {
		return nil, fmt.Errorf("select auth state state=%s: %w", state, err)
	}
	return &entity.AuthState{
		State:       dto.State,
		RedirectURL: dto.RedirectURL,
	}, nil
}

// Delete はstateをキーにしてStateTempを削除します。
func (r AuthRepository) Delete(state string) error {
	if _, err := r.dbMap.Exec("DELETE FROM auth_states where state=?", state); err != nil {
		return fmt.Errorf("delete state id=%s: %w", state, err)
	}
	return nil
}

type stateDTO struct {
	State       string `db:"state"`
	RedirectURL string `db:"redirect_url"`
}

type spotifyAuthDTO struct {
	UserID       string    `db:"user_id"`
	AccessToken  string    `db:"access_token"`
	RefreshToken string    `db:"refresh_token"`
	Expiry       time.Time `db:"expiry"`
}
