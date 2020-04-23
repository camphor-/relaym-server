package database

import (
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/google/go-cmp/cmp"
)

func TestUserRepository_FindByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      string
		want    *entity.User
		wantErr bool
	}{
		{
			name: "指定したidのユーザを取得できる",
			id:   "userID",
			want: &entity.User{
				ID: "userID",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &UserRepository{}
			got, err := r.FindByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			opt := cmp.AllowUnexported(entity.User{}, entity.SpotifyUser{})
			if !cmp.Equal(got, tt.want, opt) {
				t.Errorf("FindByID() diff=%v", cmp.Diff(got, tt.want, opt))
			}
		})
	}
}
