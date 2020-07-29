package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/camphor-/relaym-server/domain/mock_spotify"

	"github.com/camphor-/relaym-server/domain/mock_repository"
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/oauth2"

	"github.com/camphor-/relaym-server/usecase"
	"github.com/labstack/echo/v4"
)

func TestSessionTokenMiddleware_SetTokenToContext(t *testing.T) {
	tests := []struct {
		name               string
		sessionID          string
		prepareSessionRepo func(r *mock_repository.MockSession)
		prepareAuthRepo    func(r *mock_repository.MockAuth)
		prepareAuthCli     func(c *mock_spotify.MockAuth)
		next               echo.HandlerFunc
		wantErr            bool
		wantCode           int
	}{
		{
			name:      "DBからアクセストークンの取得に失敗すると500",
			sessionID: "sessionID",
			prepareSessionRepo: func(r *mock_repository.MockSession) {
				r.EXPECT().FindCreatorTokenBySessionID(gomock.Any(), "sessionID").Return(nil, "", errors.New("unknown error"))
			},
			prepareAuthRepo: func(r *mock_repository.MockAuth) {},
			prepareAuthCli:  func(c *mock_spotify.MockAuth) {},
			next:            nil,
			wantErr:         true,
			wantCode:        http.StatusInternalServerError,
		},
		{
			name:               "IDがセットされていないと404",
			sessionID:          "",
			prepareSessionRepo: func(r *mock_repository.MockSession) {},
			prepareAuthRepo:    func(r *mock_repository.MockAuth) {},
			prepareAuthCli:     func(c *mock_spotify.MockAuth) {},
			next:               nil,
			wantErr:            true,
			wantCode:           http.StatusNotFound,
		},
		{
			name:      "DBから取得したアクセストークンが正しくContextにセットされる",
			sessionID: "sessionID",
			prepareSessionRepo: func(r *mock_repository.MockSession) {
				r.EXPECT().FindCreatorTokenBySessionID(gomock.Any(), "sessionID").Return(&oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
				}, "userID", nil)
			},
			prepareAuthRepo: func(r *mock_repository.MockAuth) {},
			prepareAuthCli:  func(c *mock_spotify.MockAuth) {},
			next: func(c echo.Context) error {
				userID, ok := service.GetCreatorIDFromContext(c.Request().Context())
				if !ok {
					t.Errorf("CreatorTokenMiddleware.SetCreatorTokenToContext() userID not found in context")
				}
				if userID != "userID" {
					t.Errorf("CreatorTokenMiddleware.SetCreatorTokenToContext() userID %s, but want %s", userID, "userID")
				}
				token, ok := service.GetTokenFromContext(c.Request().Context())
				if !ok {
					t.Errorf("CreatorTokenMiddleware.SetCreatorTokenToContext() token not found in context")
				}
				want := &oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				opt := cmpopts.IgnoreUnexported(oauth2.Token{})
				if !cmp.Equal(want, token, opt) {
					t.Errorf("CreatorTokenMiddleware.SetCreatorTokenToContext() token diff = %s", cmp.Diff(want, token, opt))
				}
				return nil
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
		{
			name:      "アクセストークンの有効期限が切れているときに更新処理が走って正しく新しいトークンが保存される",
			sessionID: "sessionID",
			prepareSessionRepo: func(r *mock_repository.MockSession) {
				r.EXPECT().FindCreatorTokenBySessionID(gomock.Any(), "sessionID").Return(&oauth2.Token{
					AccessToken:  "access_token",
					TokenType:    "Bearer",
					RefreshToken: "refresh_token",
					Expiry:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
				}, "userID", nil)
			},
			prepareAuthRepo: func(r *mock_repository.MockAuth) {
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
				userID, ok := service.GetCreatorIDFromContext(c.Request().Context())
				if !ok {
					t.Errorf("CreatorTokenMiddleware.SetCreatorTokenToContext() userID not found in context")
				}
				if userID != "userID" {
					t.Errorf("CreatorTokenMiddleware.SetCreatorTokenToContext() userID %s, but want %s", userID, "userID")
				}
				token, ok := service.GetTokenFromContext(c.Request().Context())
				if !ok {
					t.Errorf("CreatorTokenMiddleware.SetCreatorTokenToContext() token not found in context")
				}
				want := &oauth2.Token{
					AccessToken:  "new_access_token",
					TokenType:    "Bearer",
					RefreshToken: "new_refresh_token",
					Expiry:       time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				opt := cmpopts.IgnoreUnexported(oauth2.Token{})
				if !cmp.Equal(want, token, opt) {
					t.Errorf("CreatorTokenMiddleware.SetCreatorTokenToContext() token diff = %s", cmp.Diff(want, token, opt))
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
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/sessions/:id/search")
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			sessionRepo := mock_repository.NewMockSession(ctrl)
			tt.prepareSessionRepo(sessionRepo)
			authRepo := mock_repository.NewMockAuth(ctrl)
			tt.prepareAuthRepo(authRepo)
			authCli := mock_spotify.NewMockAuth(ctrl)
			tt.prepareAuthCli(authCli)

			m := &CreatorTokenMiddleware{uc: usecase.NewAuthUseCase(authCli, nil, authRepo, nil, sessionRepo)}
			err := m.SetCreatorTokenToContext(tt.next)(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatorTokenMiddleware.SetCreatorTokenToContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("CreatorTokenMiddleware.SetCreatorTokenToContext() code = %d, want = %d", rec.Code, tt.wantCode)
				return
			}
		})
	}
}
