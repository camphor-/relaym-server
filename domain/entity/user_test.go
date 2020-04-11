package entity

import "testing"

func TestUser_IsPremium(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user *User
		want bool
	}{
		{
			name: "プレミアムユーザ",
			user: &User{
				SpotifyUser: SpotifyUser{
					product: "premium",
				},
			},
			want: true,
		},
		{
			name: "一般ユーザ",
			user: &User{
				SpotifyUser: SpotifyUser{
					product: "free",
				},
			},
			want: false,
		},
		{
			name: "不明なアカウントタイプを持つユーザ",
			user: &User{
				SpotifyUser: SpotifyUser{
					product: "unknown",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.IsPremium(); got != tt.want {
				t.Errorf("IsPremium() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_SpotifyURI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user *User
		want string
	}{
		{
			name: "正しくURIが構築できる",
			user: &User{SpotifyUser: SpotifyUser{id: "spotifyUserID"}},
			want: "spotify:user:spotifyUserID",
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
