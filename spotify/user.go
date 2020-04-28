package spotify

import (
	"context"
	"fmt"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/zmb3/spotify"
)

// GetMe は自分の情報をSpotify APIから取得します。
func (c *Client) GetMe(ctx context.Context) (*entity.SpotifyUser, error) {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("token not found")
	}
	cli := c.auth.NewClient(token)
	user, err := cli.CurrentUser()
	if err != nil {
		return nil, fmt.Errorf("get private user through spotiry api: %w", err)
	}
	return &entity.SpotifyUser{
		SpotifyUserID: user.ID,
		DisplayName:   user.DisplayName,
		Product:       user.Product,
	}, nil
}

// GetActiveDevices はSpotify APIを通して、ログインしているユーザがSpotifyを起動している端末を取得できます
func (c *Client) GetActiveDevices(ctx context.Context) ([]*entity.Device, error) {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("token not found")
	}
	cli := c.auth.NewClient(token)
	devices, err := cli.PlayerDevices()
	if err != nil {
		return nil, fmt.Errorf("PlayerDevices information about available devices for the current user: %w", err)
	}
	return c.toDevices(devices), nil
}

func (c *Client) toDevices(resultDevices []spotify.PlayerDevice) []*entity.Device {
	devices := make([]*entity.Device, len(resultDevices))

	for i, rd := range resultDevices {
		devices[i] = &entity.Device{
			ID:           rd.ID.String(),
			IsRestricted: rd.Restricted,
			Name:         rd.Name,
		}
	}

	return devices
}
