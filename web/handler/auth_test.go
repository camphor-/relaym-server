package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/camphor-/relaym-server/usecase"

	"github.com/labstack/echo/v4"
)

func TestAuthHandler_Callback(t *testing.T) {

	tests := []struct {
		name           string
		prepareQueryFn func() url.Values
		frontendURL    string
		wantCode       int
		wantErr        bool
		wantErrQuery   string
	}{
		{
			name: "ユーザが認可を拒否したとき",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("error", "access_denied")
				q.Set("state", "state")
				return q
			},
			frontendURL:  "relaym.local:3030",
			wantCode:     http.StatusFound,
			wantErr:      false,
			wantErrQuery: "spotifyAuthFailed",
		},
		{
			name: "stateが空のとき",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("state", "")
				q.Set("code", "code")
				return q
			},
			frontendURL:  "relaym.local:3030",
			wantCode:     http.StatusFound,
			wantErr:      false,
			wantErrQuery: "spotifyAuthFailed",
		},
		{
			name: "codeが空のとき",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("state", "state")
				q.Set("code", "")
				return q
			},
			frontendURL:  "relaym.local:3030",
			wantCode:     http.StatusFound,
			wantErr:      false,
			wantErrQuery: "spotifyAuthFailed",
		},
		{
			name: "ユーザが正しく認可したとき",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("state", "state")
				q.Set("code", "code")
				return q
			},
			frontendURL:  "relaym.local:3030",
			wantCode:     http.StatusFound,
			wantErr:      false,
			wantErrQuery: "",
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

			// TODO モックは自動生成したい
			uc := usecase.NewAuthUseCase(&fakeSpotifyAuth{}, &fakeSpotifyUser{}, &fakeAuthRepository{}, &fakeUserRepository{})
			h := &AuthHandler{
				authUC:      uc,
				frontendURL: tt.frontendURL,
			}
			if err := h.Callback(c); (err != nil) != tt.wantErr {
				t.Errorf("Callback() error = %v, wantErr %v", err, tt.wantErr)
			}
			if rec.Code != tt.wantCode {
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
