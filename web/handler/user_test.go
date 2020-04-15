package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/google/go-cmp/cmp"
	"github.com/labstack/echo/v4"
)

func TestUserHandler_GetMe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    *userRes
		wantErr bool
	}{
		{
			name: "正しくユーザが取得できる",
			want: &userRes{
				ID:          "userID",
				URI:         "spotify:user:", // TODO : Spotifyの情報も正しく取ってこれるようにする
				DisplayName: "",
				IsPremium:   false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// TODO モックは自動生成したい
			uc := usecase.NewUserUseCase(&fakeUserRepository{})
			h := &UserHandler{userUC: uc}
			if err := h.GetMe(c); (err != nil) != tt.wantErr {
				t.Errorf("GetMe() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := &userRes{}
			err := json.Unmarshal(rec.Body.Bytes(), got)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("GetMe() diff = %v", cmp.Diff(got, tt.want))
			}
		})
	}
}

type fakeUserRepository struct {
}

func (f fakeUserRepository) FindByID(id string) (*entity.User, error) {
	return entity.NewUser(id), nil
}
