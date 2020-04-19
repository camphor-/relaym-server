package database

import (
	"fmt"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/go-gorp/gorp/v3"
)

var _ repository.Auth = &AuthRepository{}

// AuthRepository は repository.AuthRepository を満たす構造体です
type AuthRepository struct {
	dbMap *gorp.DbMap
}

// NewAuthRepository はAuthRepositoryのポインタを生成する関数です
func NewAuthRepository(dbMap *gorp.DbMap) *AuthRepository {
	dbMap.AddTableWithName(stateDTO{}, "auth_states")
	return &AuthRepository{dbMap: dbMap}
}

// Store はauthStateを保存します。
func (r AuthRepository) Store(authState *entity.AuthState) error {
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
