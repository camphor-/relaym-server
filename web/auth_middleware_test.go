package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/camphor-/relaym-server/config"
	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/mock_repository"
	"github.com/camphor-/relaym-server/domain/mock_spotify"
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func TestAuthMiddleware_Authenticate(t *testing.T) {

	tests := []struct {
		name            string
		prepareRequest  func(req *http.Request)
		prepareAuthRepo func(r *mock_repository.MockAuth)
		prepareAuthCli  func(c *mock_spotify.MockAuth)
		next            echo.HandlerFunc
		wantErr         bool
		wantCode        int
	}{
		{
			name: "セッションがクッキーに存在しないと401",
			prepareRequest: func(req *http.Request) {
			},
			prepareAuthRepo: func(r *mock_repository.MockAuth) {},
			prepareAuthCli:  func(c *mock_spotify.MockAuth) {},
			next:            nil,
			wantErr:         true,
			wantCode:        http.StatusUnauthorized,
		},
		{
			name: "DBからセッションの取得に失敗すると401",
			prepareRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:     "session",
					Value:    "sessionID",
					Path:     "/",
					MaxAge:   60 * 60 * 24 * 7,
					Secure:   !config.IsLocal(),
					HttpOnly: true,
					SameSite: http.SameSiteNoneMode,
				})
			},
			prepareAuthRepo: func(r *mock_repository.MockAuth) {
				r.EXPECT().GetUserIDFromSession("sessionID").Return("", errors.New("unknown error"))
			},
			prepareAuthCli: func(c *mock_spotify.MockAuth) {},
			next:           nil,
			wantErr:        true,
			wantCode:       http.StatusUnauthorized,
		},
		{
			name: "DBからアクセストークンの取得に失敗すると401",
			prepareRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:     "session",
					Value:    "sessionID",
					Path:     "/",
					MaxAge:   60 * 60 * 24 * 7,
					Secure:   !config.IsLocal(),
					HttpOnly: true,
					SameSite: http.SameSiteNoneMode,
				})
			},
			prepareAuthRepo: func(r *mock_repository.MockAuth) {
				r.EXPECT().GetUserIDFromSession("sessionID").Return("userID", nil)
				r.EXPECT().GetTokenByUserID("userID").Return(nil, errors.New("unknown error"))
			},
			prepareAuthCli: func(c *mock_spotify.MockAuth) {},
			next:           nil,
			wantErr:        true,
			wantCode:       http.StatusInternalServerError,
		},
		{
			name: "DBにアクセストークンが存在しないと401",
			prepareRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:     "session",
					Value:    "sessionID",
					Path:     "/",
					MaxAge:   60 * 60 * 24 * 7,
					Secure:   !config.IsLocal(),
					HttpOnly: true,
					SameSite: http.SameSiteNoneMode,
				})
			},
			prepareAuthRepo: func(r *mock_repository.MockAuth) {
				r.EXPECT().GetUserIDFromSession("sessionID").Return("userID", nil)
				r.EXPECT().GetTokenByUserID("userID").Return(nil, entity.ErrTokenNotFound)
			},
			prepareAuthCli: func(c *mock_spotify.MockAuth) {},
			next:           nil,
			wantErr:        true,
			wantCode:       http.StatusUnauthorized,
		},
		{
			name: "DBから取得したアクセストークンを正しくContextにセットされる",
			prepareRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:     "session",
					Value:    "sessionID",
					Path:     "/",
					MaxAge:   60 * 60 * 24 * 7,
					Secure:   !config.IsLocal(),
					HttpOnly: true,
					SameSite: http.SameSiteNoneMode,
				})
			},
			prepareAuthRepo: func(r *mock_repository.MockAuth) {
				r.EXPECT().GetUserIDFromSession("sessionID").Return("userID", nil)
				r.EXPECT().GetTokenByUserID("userID").Return(&oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			prepareAuthCli: func(c *mock_spotify.MockAuth) {},
			next: func(c echo.Context) error {
				userID, ok := service.GetUserIDFromContext(c.Request().Context())
				if !ok {
					t.Errorf("AuthMiddleware.Authenticate() userID not found in context")
				}
				if userID != "userID" {
					t.Errorf("AuthMiddleware.Authenticate() userID %s, but want %s", userID, "userID")
				}
				token, ok := service.GetTokenFromContext(c.Request().Context())
				if !ok {
					t.Errorf("AuthMiddleware.Authenticate() token not found in context")
				}
				want := &oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				opt := cmpopts.IgnoreUnexported(oauth2.Token{})
				if !cmp.Equal(want, token, opt) {
					t.Errorf("AuthMiddleware.Authenticate() token diff = %s", cmp.Diff(want, token, opt))
				}
				return nil
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
		{
			name: "アクセストークンの有効期限が切れているときは更新処理が走って正しく新しいトークンが保存される",
			prepareRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:     "session",
					Value:    "sessionID",
					Path:     "/",
					MaxAge:   60 * 60 * 24 * 7,
					Secure:   !config.IsLocal(),
					HttpOnly: true,
					SameSite: http.SameSiteNoneMode,
				})
			},
			prepareAuthRepo: func(r *mock_repository.MockAuth) {
				r.EXPECT().GetUserIDFromSession("sessionID").Return("userID", nil)
				r.EXPECT().GetTokenByUserID("userID").Return(&oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
				r.EXPECT().StoreORUpdateToken("userID", &oauth2.Token{
					AccessToken:  "new_access_token",
					TokenType:    "Bearer",
					RefreshToken: "new_refresh_token",
					Expiry:       time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
				}).Return(nil)
			},
			prepareAuthCli: func(c *mock_spotify.MockAuth) {
				c.EXPECT().Refresh(&oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
				}).Return(&oauth2.Token{
					AccessToken:  "new_access_token",
					TokenType:    "Bearer",
					RefreshToken: "new_refresh_token",
					Expiry:       time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			next: func(c echo.Context) error {
				userID, ok := service.GetUserIDFromContext(c.Request().Context())
				if !ok {
					t.Errorf("AuthMiddleware.Authenticate() userID not found in context")
				}
				if userID != "userID" {
					t.Errorf("AuthMiddleware.Authenticate() userID %s, but want %s", userID, "userID")
				}
				token, ok := service.GetTokenFromContext(c.Request().Context())
				if !ok {
					t.Errorf("AuthMiddleware.Authenticate() token not found in context")
				}
				want := &oauth2.Token{
					AccessToken:  "new_access_token",
					TokenType:    "Bearer",
					RefreshToken: "new_refresh_token",
					Expiry:       time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				opt := cmpopts.IgnoreUnexported(oauth2.Token{})
				if !cmp.Equal(want, token, opt) {
					t.Errorf("AuthMiddleware.Authenticate() token diff = %s", cmp.Diff(want, token, opt))
				}
				return nil
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			tt.prepareRequest(req)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authRepo := mock_repository.NewMockAuth(ctrl)
			tt.prepareAuthRepo(authRepo)
			authCli := mock_spotify.NewMockAuth(ctrl)
			tt.prepareAuthCli(authCli)

			m := &AuthMiddleware{uc: usecase.NewAuthUseCase(authCli, nil, authRepo, nil)}
			err := m.Authenticate(tt.next)(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthMiddleware.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("AuthMiddleware.Authenticate() code = %d, want = %d", rec.Code, tt.wantCode)
			}
		})
	}
}
