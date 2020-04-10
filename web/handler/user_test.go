package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/google/go-cmp/cmp"
)

func TestUserHandler_GetMe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    *userJSON
		wantErr bool
	}{
		{
			name:    "",
			want:    &userJSON{},
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

			h := &UserHandler{}
			if err := h.GetMe(c); (err != nil) != tt.wantErr {
				t.Errorf("GetMe() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := &userJSON{}
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
