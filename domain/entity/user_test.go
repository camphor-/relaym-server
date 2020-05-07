package entity

import "testing"

func TestUser_SpotifyURI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user *User
		want string
	}{
		{
			name: "正しくURIが構築できる",
			user: &User{SpotifyUserID: "UserID"},
			want: "spotify:user:UserID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.SpotifyURI(); got != tt.want {
				t.Errorf("SpotifyURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpotifyUser_Premium(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		su   *SpotifyUser
		want bool
	}{
		{
			name: "プレミアム会員",
			su: &SpotifyUser{
				Product: "premium",
			},
			want: true,
		},
		{
			name: "一般会員",
			su: &SpotifyUser{
				Product: "free",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.su.Premium(); got != tt.want {
				t.Errorf("Premium() = %v, want %v", got, tt.want)
			}
		})
	}
}
