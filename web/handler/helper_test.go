package handler

import (
	"context"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"golang.org/x/oauth2"
)

type fakeSpotifyUser struct{}

func (f fakeSpotifyUser) GetMe(ctx context.Context) (*entity.SpotifyUser, error) {
	return &entity.SpotifyUser{}, nil
}

type fakeSpotifyAuth struct{}

func (f fakeSpotifyAuth) GetAuthURL(state string) string {
	return ""
}

func (f fakeSpotifyAuth) Exchange(code string) (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken:  "token",
		TokenType:    "Bearer",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(1 * time.Hour),
	}, nil
}

type fakeAuthRepository struct{}

func (f fakeAuthRepository) StoreORUpdateToken(spotifyUserID string, token *oauth2.Token) error {
	return nil
}

func (f fakeAuthRepository) GetTokenBySpotifyUserID(spotifyUserID string) (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken:  "access_token",
		TokenType:    "Bearer",
		RefreshToken: "refresh_token",
		Expiry:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}, nil
}

func (f fakeAuthRepository) StoreState(state *entity.AuthState) error {
	return nil
}

func (f fakeAuthRepository) FindStateByState(state string) (*entity.AuthState, error) {
	return &entity.AuthState{
		State:       "state",
		RedirectURL: "https://example.com",
	}, nil
}

func (f fakeAuthRepository) Delete(state string) error {
	return nil
}

type fakeUserRepository struct {
}

func (f fakeUserRepository) FindBySpotifyUserID(spotifyUserID string) (*entity.User, error) {
	return &entity.User{}, nil
}

func (f fakeUserRepository) Store(user *entity.User) error {
	return nil
}

func (f fakeUserRepository) FindByID(id string) (*entity.User, error) {
	return &entity.User{ID: id}, nil
}
