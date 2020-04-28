package spotify

import (
	"context"
	"fmt"
	//"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/service"
)

// GetActiveDevices はSpotify APIを通して、ログインしているユーザがSpotifyを起動している端末を取得できます
func (c *Client) Search(ctx context.Context, q string) ([]*entity.Device, error) {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("token not found")
	}

	cli := c.auth.NewClient(token)
	result, err := cli.PlayerDevices()
	if err != nil {
		return nil, fmt.Errorf("search q=%s: %w", q, err)
	}
	return nil, nil
}
