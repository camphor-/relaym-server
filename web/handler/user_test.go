package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/mock_repository"
	"github.com/camphor-/relaym-server/domain/mock_spotify"
	"github.com/camphor-/relaym-server/usecase"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/labstack/echo/v4"
)

func TestUserHandler_GetMe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		userID              string
		prepareMockUserRepo func(repo *mock_repository.MockUser)
		prepareMockUserCli  func(cli *mock_spotify.MockUser)
		want                *userRes
		wantErr             bool
		wantCode            int
	}{
		{
			name:   "正しくユーザが取得できる",
			userID: "userID",
			prepareMockUserRepo: func(repo *mock_repository.MockUser) {
				repo.EXPECT().FindByID("userID").Return(&entity.User{
					ID:            "userID",
					SpotifyUserID: "spotify_user_id",
					DisplayName:   "display_name",
				}, nil)
			},
			prepareMockUserCli: func(cli *mock_spotify.MockUser) {
				cli.EXPECT().GetMe(gomock.Any()).Return(&entity.SpotifyUser{
					SpotifyUserID: "spotify_user_id",
					DisplayName:   "spotify_display_name",
					Product:       "premium",
				}, nil)
			},
			want: &userRes{
				ID:          "userID",
				URI:         "spotify:user:spotify_user_id",
				DisplayName: "display_name",
				IsPremium:   true,
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
		{
			name:   "DBからユーザの取得に失敗したときはInternalServerError",
			userID: "userID",
			prepareMockUserRepo: func(repo *mock_repository.MockUser) {
				repo.EXPECT().FindByID("userID").Return(nil, errors.New("unknown error"))
			},
			prepareMockUserCli: func(cli *mock_spotify.MockUser) {},
			want:               nil,
			wantErr:            true,
			wantCode:           http.StatusInternalServerError,
		},
		{
			name:   "SpotifyAPIからユーザの取得に失敗したときはInternalServerError",
			userID: "userID",
			prepareMockUserRepo: func(repo *mock_repository.MockUser) {
				repo.EXPECT().FindByID("userID").Return(&entity.User{
					ID:            "userID",
					SpotifyUserID: "spotify_user_id",
					DisplayName:   "display_name",
				}, nil)
			},
			prepareMockUserCli: func(cli *mock_spotify.MockUser) {
				cli.EXPECT().GetMe(gomock.Any()).Return(nil, errors.New("unknown error"))
			},
			want:     nil,
			wantErr:  true,
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c = setToContext(c, tt.userID, nil)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := mock_repository.NewMockUser(ctrl)
			tt.prepareMockUserRepo(repo)
			cli := mock_spotify.NewMockUser(ctrl)
			tt.prepareMockUserCli(cli)
			uc := usecase.NewUserUseCase(cli, repo)
			h := &UserHandler{userUC: uc}

			err := h.GetMe(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMe() error = %v, wantErr %v", err, tt.wantErr)
			}
			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("GetMe() code = %d, want = %d", rec.Code, tt.wantCode)
			}

			if !tt.wantErr {
				got := &userRes{}
				err := json.Unmarshal(rec.Body.Bytes(), got)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, tt.want) {
					t.Errorf("GetMe() diff = %v", cmp.Diff(got, tt.want))
				}
			}
		})
	}
}
