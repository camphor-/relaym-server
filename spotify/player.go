package spotify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/camphor-/relaym-server/log"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/service"

	"github.com/zmb3/spotify/v2"
)

// CurrentlyPlaying は現在の再生状況を取得するAPIです。
func (c *Client) CurrentlyPlaying(ctx context.Context) (*entity.CurrentPlayingInfo, error) {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return nil, errors.New("token not found")
	}
	cli := spotify.New(c.auth.Client(ctx, token))
	ps, err := cli.PlayerState(ctx)
	if convErr := c.convertPlayerError(err); convErr != nil {
		return nil, fmt.Errorf("spotify api: currently playing: %w", convErr)
	}
	return &entity.CurrentPlayingInfo{
		Playing:  ps.Playing,
		Progress: time.Duration(ps.Progress) * time.Millisecond,
		Track:    c.toTrack(ps.Item),
		Device:   c.toDevice(ps.Device),
	}, nil
}

func (c *Client) toDevice(device spotify.PlayerDevice) *entity.Device {
	return &entity.Device{
		ID:           string(device.ID),
		IsRestricted: device.Restricted,
		Name:         device.Name,
	}
}

// GoNextTrack はユーザーが現在再生している曲を1曲skipします。
// APIが非同期で処理がされるため、リクエストが返ってきてもskipが完了しているとは限りません。
// プレミアム会員必須
func (c *Client) GoNextTrack(ctx context.Context, deviceID string) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := spotify.New(c.auth.Client(ctx, token))

	opt := &spotify.PlayOptions{DeviceID: nil}
	if deviceID != "" {
		spotifyID := spotify.ID(deviceID)
		opt = &spotify.PlayOptions{DeviceID: &spotifyID}
	}

	err := cli.NextOpt(ctx, opt)
	if convErr := c.convertPlayerError(err); convErr != nil {
		return fmt.Errorf("spotify api: next: %w", convErr)
	}

	return nil
}

// DeleteAllTracksInQueue はユーザーのSpotifyに積まれている「次に再生される曲」「再生待ち」を全てskipします。
// プレミアム会員必須
func (c *Client) DeleteAllTracksInQueue(ctx context.Context, deviceID string, trackURI string) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := spotify.New(c.auth.Client(ctx, token))

	// PlayWithTracksで「再生待ち」を0曲にする
	if err := c.PlayWithTracks(ctx, deviceID, []string{trackURI}); err != nil {
		return fmt.Errorf("call play api with tracks %v: %w", trackURI, err)
	}

	opt := &spotify.PlayOptions{DeviceID: nil}
	if deviceID != "" {
		spotifyID := spotify.ID(deviceID)
		opt = &spotify.PlayOptions{DeviceID: &spotifyID}
	}

	skipOnceTime := 3
	sleepTime := 300 * time.Millisecond
	for i := 1; ; i++ {
		err := cli.NextOpt(ctx, opt)
		if convErr := c.convertPlayerError(err); convErr != nil {
			return fmt.Errorf("spotify api: next: %w", convErr)
		}

		if i%skipOnceTime == 0 {
			// SpotifyAPIを叩いてからSpotifyが曲をskipするのに時間がかかるため余計にAPIを叩かないように調節
			time.Sleep(sleepTime)
			cpi, err := c.CurrentlyPlaying(ctx)
			if err != nil {
				return fmt.Errorf("spotify api: CurrentlyPlaying: %w", err)
			}

			if !cpi.Playing {
				break
			}
		}
	}
	return nil
}

// Play は曲を再生し始めるか現在再生途中の曲の再生を再開するAPIです。deviceIDが空の場合はデフォルトのデバイスで再生されます。
// APIが非同期で処理がされるため、リクエストが返ってきても再生が開始しているとは限りません。
// 設定が反映されたか確認するには CurrentlyPlaying() を叩く必要があります。
// プレミアム会員必須
func (c *Client) Play(ctx context.Context, deviceID string) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := spotify.New(c.auth.Client(ctx, token))

	opt := &spotify.PlayOptions{DeviceID: nil}
	if deviceID != "" {
		spotifyID := spotify.ID(deviceID)
		opt = &spotify.PlayOptions{DeviceID: &spotifyID}
	}

	err := cli.PlayOpt(ctx, opt)
	if convErr := c.convertPlayerError(err); convErr != nil {
		return fmt.Errorf("spotify api: play or resume: %w", convErr)
	}
	return nil
}

// PlayWithTracks は曲を指定して曲を再生し始めるAPIです。deviceIDが空の場合はデフォルトのデバイスで再生されます。
// APIが非同期で処理がされるため、リクエストが返ってきても再生が開始しているとは限りません。
// 設定が反映されたか確認するには CurrentlyPlaying() を叩く必要があります。
// プレミアム会員必須
func (c *Client) PlayWithTracks(ctx context.Context, deviceID string, trackURIs []string) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := spotify.New(c.auth.Client(ctx, token))

	opt := &spotify.PlayOptions{DeviceID: nil, URIs: c.toURIs(trackURIs)}
	if deviceID != "" {
		spotifyID := spotify.ID(deviceID)
		opt = &spotify.PlayOptions{DeviceID: &spotifyID, URIs: c.toURIs(trackURIs)}
	}

	err := cli.PlayOpt(ctx, opt)
	if convErr := c.convertPlayerError(err); convErr != nil {
		return fmt.Errorf("spotify api: play or resume: %w", convErr)
	}
	return nil
}

// PlayWithTracksAndPosition は指定した曲を、指定した位置から再生を始めるAPIです。deviceIDが空の場合はデフォルトのデバイスで再生されます。
// APIが非同期で処理がされるため、リクエストが返ってきても再生が開始しているとは限りません。
// 設定が反映されたか確認するには CurrentlyPlaying() を叩く必要があります。
// プレミアム会員必須
func (c *Client) PlayWithTracksAndPosition(ctx context.Context, deviceID string, trackURIs []string, position time.Duration) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := spotify.New(c.auth.Client(ctx, token))

	opt := &spotify.PlayOptions{DeviceID: nil, URIs: c.toURIs(trackURIs), PositionMs: int(position.Milliseconds())}
	if deviceID != "" {
		spotifyID := spotify.ID(deviceID)
		opt = &spotify.PlayOptions{DeviceID: &spotifyID, URIs: c.toURIs(trackURIs)}
	}

	err := cli.PlayOpt(ctx, opt)
	if convErr := c.convertPlayerError(err); convErr != nil {
		return fmt.Errorf("spotify api: play or resume: %w", convErr)
	}
	return nil
}

// Pause は再生を一時停止します。deviceIDが空の場合はデフォルトのデバイスで再生されます。
// APIが非同期で処理がされるため、リクエストが返ってきても再生が一時停止されているとは限りません。
// 設定が反映されたか確認するには CurrentlyPlaying() を叩く必要があります。
// プレミアム会員必須
func (c *Client) Pause(ctx context.Context, deviceID string) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := spotify.New(c.auth.Client(ctx, token))

	opt := &spotify.PlayOptions{DeviceID: nil}
	if deviceID != "" {
		spotifyID := spotify.ID(deviceID)
		opt = &spotify.PlayOptions{DeviceID: &spotifyID}
	}
	err := cli.PauseOpt(ctx, opt)
	if convErr := c.convertPlayerError(err); convErr != nil {
		return fmt.Errorf("spotify api: pause: %w", convErr)
	}
	return nil
}

// Enqueue は曲を「次に再生される曲」に追加するAPIです。deviceIDが空の場合はデフォルトのデバイスで再生されます。
// APIが非同期で処理がされるため、リクエストが返ってきても曲の追加が完了しているとは限りません。
// 設定が反映されたか確認するには CurrentlyPlaying() を叩く必要があります。
// プレミアム会員必須
func (c *Client) Enqueue(ctx context.Context, trackURI string, deviceID string) error {
	trackID := strings.Replace(trackURI, "spotify:track:", "", 1)

	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := spotify.New(c.auth.Client(ctx, token))

	opt := &spotify.PlayOptions{DeviceID: nil}
	if deviceID != "" {
		spotifyID := spotify.ID(deviceID)
		opt = &spotify.PlayOptions{DeviceID: &spotifyID}
	}
	err := cli.QueueSongOpt(ctx, spotify.ID(trackID), opt)
	if convErr := c.convertPlayerError(err); convErr != nil {
		return fmt.Errorf("spotify api: add queue: %w", convErr)
	}
	return nil
}

// SetRepeatMode はリピートモードの設定を変更するAPIです。
// APIが非同期で処理がされるため、リクエストが返ってきてもリピートモードの設定が完了しているとは限りません。
// 設定が反映されたか確認するには CurrentlyPlaying() を叩く必要があります。
// プレミアム会員必須
func (c *Client) SetRepeatMode(ctx context.Context, on bool, deviceID string) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := spotify.New(c.auth.Client(ctx, token))

	state := "off"
	if on {
		state = "context"
	}

	opt := &spotify.PlayOptions{DeviceID: nil}
	if deviceID != "" {
		spotifyID := spotify.ID(deviceID)
		opt = &spotify.PlayOptions{DeviceID: &spotifyID}
	}

	if err := cli.RepeatOpt(ctx, state, opt); c.convertPlayerError(err) != nil {
		return fmt.Errorf("spotify api: set repeat mode: %w", c.convertPlayerError(err))
	}
	return nil
}

// SetShuffleMode はシャッフルモードの設定を変更するAPIです。
// APIが非同期で処理がされるため、リクエストが返ってきてもシャッフルモードの設定が完了しているとは限りません。
// 設定が反映されたか確認するには CurrentlyPlaying() を叩く必要があります。
// プレミアム会員必須
func (c *Client) SetShuffleMode(ctx context.Context, on bool, deviceID string) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := spotify.New(c.auth.Client(ctx, token))

	opt := &spotify.PlayOptions{DeviceID: nil}
	if deviceID != "" {
		spotifyID := spotify.ID(deviceID)
		opt = &spotify.PlayOptions{DeviceID: &spotifyID}
	}
	if err := cli.ShuffleOpt(ctx, on, opt); c.convertPlayerError(err) != nil {
		return fmt.Errorf("spotify api: set repeat mode: %w", c.convertPlayerError(err))
	}
	return nil
}

func (c *Client) convertPlayerError(err error) error {
	logger := log.New()
	if e, ok := err.(spotify.Error); ok {
		switch {
		case e.Status == http.StatusForbidden && strings.Contains(e.Message, "Restriction violated"):
			// https://github.com/spotify/web-api/issues/1205
			logger.Infoj(map[string]interface{}{
				"message":     "already in the ideal state",
				"apiResponse": e.Message,
			})
			return nil
		case e.Status == http.StatusForbidden:
			return fmt.Errorf("%s: %w", e.Message, entity.ErrNonPremium)
		case e.Status == http.StatusNotFound || (e.Status == http.StatusInternalServerError && strings.Contains(e.Message, "Server error")):
			return fmt.Errorf("%s: %w", e.Message, entity.ErrActiveDeviceNotFound)
		}
	}
	return err
}

func (c *Client) toURIs(uris []string) []spotify.URI {
	sURIs := make([]spotify.URI, len(uris))
	for i := 0; i < len(uris); i++ {
		sURIs[i] = spotify.URI(uris[i])
	}
	return sURIs
}
