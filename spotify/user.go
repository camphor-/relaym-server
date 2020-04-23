package spotify

import (
	"context"
	"fmt"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/service"
)

// GetMe は自分の情報をSpotify APIから取得します。
// 返すUserのidは存在しないので注意
func (c *Client) GetMe(ctx context.Context) (*entity.SpotifyUser, error) {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("token not found")
	}
	cli := c.auth.NewClient(token)
	user, err := cli.CurrentUser()
	if err != nil {
		return nil, fmt.Errorf("get private user though spotiry api: %w", err)
	}
	return &entity.SpotifyUser{
		SpotifyUserID: user.ID,
		DisplayName:   user.DisplayName,
		Product:       user.Product,
	}, nil
}
