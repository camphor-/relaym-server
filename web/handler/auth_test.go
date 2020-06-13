package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/mock_repository"
	"github.com/camphor-/relaym-server/domain/mock_spotify"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func TestAuthHandler_Login(t *testing.T) {

	tests := []struct {
		name                  string
		frontendURL           string
		prepareQueryFn        func() url.Values
		prepareMockAuthRepoFn func(mock *mock_repository.MockAuth)
		prepareMockAuthSpoFn  func(mock *mock_spotify.MockAuth)
		wantErr               bool
		wantCode              int
		wantErrQuery          string
	}{
		{
			name:        "redirect_urlが存在するとき正しく動作する",
			frontendURL: "http://relaym.local:3030",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("redirect_url", "http://relaym.local:3030/redirect")
				return q
			},
			prepareMockAuthRepoFn: func(mock *mock_repository.MockAuth) {
				mock.EXPECT().StoreState(gomock.Any()).Return(nil)
			},
			prepareMockAuthSpoFn: func(mock *mock_spotify.MockAuth) {
				mock.EXPECT().GetAuthURL(gomock.Any()).Return("https://spotify.com/hogehoge")
			},
			wantErr:      false,
			wantCode:     http.StatusFound,
			wantErrQuery: "",
		},
		{
			name:        "クエリにredirect_urlがなくてもエラーにならない",
			frontendURL: "http://relaym.local:3030",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				return q
			},
			prepareMockAuthRepoFn: func(mock *mock_repository.MockAuth) {
				mock.EXPECT().StoreState(gomock.Any()).Return(nil)
			},
			prepareMockAuthSpoFn: func(mock *mock_spotify.MockAuth) {
				mock.EXPECT().GetAuthURL(gomock.Any()).Return("https://spotify.com/hogehoge")
			},
			wantErr:      false,
			wantCode:     http.StatusFound,
			wantErrQuery: "",
		},
		{
			name:        "Stateの保存に失敗するとエラーになる",
			frontendURL: "http://relaym.local:3030",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("redirect_url", "http://relaym.local:3030/redirect")
				return q
			},
			prepareMockAuthRepoFn: func(mock *mock_repository.MockAuth) {
				mock.EXPECT().StoreState(gomock.Any()).Return(errors.New("unknown error"))
			},
			prepareMockAuthSpoFn: func(mock *mock_spotify.MockAuth) {},
			wantErr:              false,
			wantCode:             http.StatusFound,
			wantErrQuery:         "spotifyAuthFailed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			q := tt.prepareQueryFn()
			req := httptest.NewRequest(http.MethodPost, "/?"+q.Encode(), nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockAuthRepo := mock_repository.NewMockAuth(ctrl)
			tt.prepareMockAuthRepoFn(mockAuthRepo)
			mockAuthSpo := mock_spotify.NewMockAuth(ctrl)
			tt.prepareMockAuthSpoFn(mockAuthSpo)
			h := &AuthHandler{
				authUC:      usecase.NewAuthUseCase(mockAuthSpo, nil, mockAuthRepo, nil, nil),
				frontendURL: tt.frontendURL,
			}

			err := h.Login(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("Login() code = %d, want = %d", rec.Code, tt.wantCode)
			}

			u, err := rec.Result().Location()
			if tt.wantCode == http.StatusFound && (u == nil || err != nil) {
				t.Fatal(err)
				return
			}

			if u.Query().Get("err") != tt.wantErrQuery {
				t.Errorf("Callback() err query = %s, want = %s", u.Query().Get("err"), tt.wantErrQuery)
			}
		})
	}
}

func TestAuthHandler_Callback(t *testing.T) {
	tests := []struct {
		name string

		prepareQueryFn        func() url.Values
		prepareMockAuthRepoFn func(mock *mock_repository.MockAuth)
		prepareMockUserRepoFn func(mock *mock_repository.MockUser)
		prepareMockAuthSpoFn  func(mock *mock_spotify.MockAuth)
		prepareMockUserSpoFn  func(mock *mock_spotify.MockUser)
		frontendURL           string
		wantCode              int
		wantErr               bool
		wantErrQuery          string
	}{
		{
			name: "ユーザが認可を拒否したとき",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("error", "access_denied")
				q.Set("state", "state")
				return q
			},
			prepareMockAuthRepoFn: func(mock *mock_repository.MockAuth) {},
			prepareMockUserRepoFn: func(mock *mock_repository.MockUser) {},
			prepareMockAuthSpoFn:  func(mock *mock_spotify.MockAuth) {},
			prepareMockUserSpoFn:  func(mock *mock_spotify.MockUser) {},
			frontendURL:           "relaym.local:3030",
			wantCode:              http.StatusFound,
			wantErr:               false,
			wantErrQuery:          "spotifyAuthFailed",
		},
		{
			name: "stateが空のとき",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("state", "")
				q.Set("code", "code")
				return q
			},
			prepareMockAuthRepoFn: func(mock *mock_repository.MockAuth) {},
			prepareMockUserRepoFn: func(mock *mock_repository.MockUser) {},
			prepareMockAuthSpoFn:  func(mock *mock_spotify.MockAuth) {},
			prepareMockUserSpoFn:  func(mock *mock_spotify.MockUser) {},
			frontendURL:           "relaym.local:3030",
			wantCode:              http.StatusFound,
			wantErr:               false,
			wantErrQuery:          "spotifyAuthFailed",
		},
		{
			name: "codeが空のとき",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("state", "state")
				q.Set("code", "")
				return q
			},
			prepareMockAuthRepoFn: func(mock *mock_repository.MockAuth) {},
			prepareMockUserRepoFn: func(mock *mock_repository.MockUser) {},
			prepareMockAuthSpoFn:  func(mock *mock_spotify.MockAuth) {},
			prepareMockUserSpoFn:  func(mock *mock_spotify.MockUser) {},
			frontendURL:           "relaym.local:3030",
			wantCode:              http.StatusFound,
			wantErr:               false,
			wantErrQuery:          "spotifyAuthFailed",
		},
		{
			name: "ユーザが正しく認可したとき",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("state", "state")
				q.Set("code", "code")
				return q
			},
			prepareMockAuthRepoFn: func(mock *mock_repository.MockAuth) {
				mock.EXPECT().FindStateByState("state").Return(&entity.AuthState{
					State:       "state",
					RedirectURL: "relaym.local:3030",
				}, nil)
				mock.EXPECT().StoreORUpdateToken(gomock.Any(), &oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				}).Return(nil)
				mock.EXPECT().StoreSession(gomock.Any(), gomock.Any()).Return(nil)
				mock.EXPECT().DeleteState("state").Return(nil)
			},
			prepareMockUserRepoFn: func(mock *mock_repository.MockUser) {
				mock.EXPECT().Store(gomock.Any()).Return(nil)
			},
			prepareMockAuthSpoFn: func(mock *mock_spotify.MockAuth) {
				mock.EXPECT().Exchange("code").Return(&oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			prepareMockUserSpoFn: func(mock *mock_spotify.MockUser) {
				mock.EXPECT().GetMe(gomock.Any()).Return(&entity.SpotifyUser{
					SpotifyUserID: "spotify_user_id",
					DisplayName:   "display_name",
					Product:       "premium",
				}, nil)
			},
			frontendURL:  "relaym.local:3030",
			wantCode:     http.StatusFound,
			wantErr:      false,
			wantErrQuery: "",
		},
		{
			name: "Stateの削除が失敗してもエラーを返さない",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("state", "state")
				q.Set("code", "code")
				return q
			},
			prepareMockAuthRepoFn: func(mock *mock_repository.MockAuth) {
				mock.EXPECT().FindStateByState("state").Return(&entity.AuthState{
					State:       "state",
					RedirectURL: "relaym.local:3030",
				}, nil)
				mock.EXPECT().StoreORUpdateToken(gomock.Any(), &oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				}).Return(nil)
				mock.EXPECT().StoreSession(gomock.Any(), gomock.Any()).Return(nil)
				mock.EXPECT().DeleteState("state").Return(errors.New("unknown error"))
			},
			prepareMockUserRepoFn: func(mock *mock_repository.MockUser) {
				mock.EXPECT().Store(gomock.Any()).Return(nil)
			},
			prepareMockAuthSpoFn: func(mock *mock_spotify.MockAuth) {
				mock.EXPECT().Exchange("code").Return(&oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			prepareMockUserSpoFn: func(mock *mock_spotify.MockUser) {
				mock.EXPECT().GetMe(gomock.Any()).Return(&entity.SpotifyUser{
					SpotifyUserID: "spotify_user_id",
					DisplayName:   "display_name",
					Product:       "premium",
				}, nil)
			},
			frontendURL:  "relaym.local:3030",
			wantCode:     http.StatusFound,
			wantErr:      false,
			wantErrQuery: "",
		},
		{
			name: "Authorizationの途中で失敗したらエラーになる",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("state", "state")
				q.Set("code", "code")
				return q
			},
			prepareMockAuthRepoFn: func(mock *mock_repository.MockAuth) {
				mock.EXPECT().FindStateByState("state").Return(nil, errors.New("unknown error"))
			},
			prepareMockUserRepoFn: func(mock *mock_repository.MockUser) {},
			prepareMockAuthSpoFn:  func(mock *mock_spotify.MockAuth) {},
			prepareMockUserSpoFn:  func(mock *mock_spotify.MockUser) {},
			frontendURL:           "relaym.local:3030",
			wantCode:              http.StatusFound,
			wantErr:               false,
			wantErrQuery:          "spotifyAuthFailed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			q := tt.prepareQueryFn()
			req := httptest.NewRequest(http.MethodPost, "/?"+q.Encode(), nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockAuthSpo := mock_spotify.NewMockAuth(ctrl)
			tt.prepareMockAuthSpoFn(mockAuthSpo)
			mockUserSpo := mock_spotify.NewMockUser(ctrl)
			tt.prepareMockUserSpoFn(mockUserSpo)
			mockAuthRepo := mock_repository.NewMockAuth(ctrl)
			tt.prepareMockAuthRepoFn(mockAuthRepo)
			mockUserRepo := mock_repository.NewMockUser(ctrl)
			tt.prepareMockUserRepoFn(mockUserRepo)

			uc := usecase.NewAuthUseCase(mockAuthSpo, mockUserSpo, mockAuthRepo, mockUserRepo, nil)
			h := &AuthHandler{
				authUC:      uc,
				frontendURL: tt.frontendURL,
			}
			err := h.Callback(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Callback() error = %v, wantErr %v", err, tt.wantErr)
			}
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("Callback() code = %d, want = %d", rec.Code, tt.wantCode)
			}

			u, err := rec.Result().Location()
			if tt.wantCode == http.StatusFound && (u == nil || err != nil) {
				t.Fatal(err)
				return
			}

			if u.Query().Get("err") != tt.wantErrQuery {
				t.Errorf("Callback() err query = %s, want = %s", u.Query().Get("err"), tt.wantErrQuery)
			}
		})
	}
}
