package spotify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/service"

	"github.com/zmb3/spotify"
)

// CurrentlyPlaying は現在の再生状況を取得するAPIです。
// TODO : 現状はなんの情報が必要か分かってないので現在再生中かどうかのみ返します。
func (c *Client) CurrentlyPlaying(ctx context.Context) (*entity.CurrentPlayingInfo, error) {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return nil, errors.New("token not found")
	}
	cli := c.auth.NewClient(token)
	cp, err := cli.PlayerCurrentlyPlaying()
	if convErr := c.convertPlayerError(err); convErr != nil {
		return nil, fmt.Errorf("spotify api: currently playing: %w", convErr)
	}
	return &entity.CurrentPlayingInfo{
		Playing:  cp.Playing,
		Progress: time.Duration(cp.Progress) * time.Millisecond,
		Track:    c.toTrack(cp.Item),
	}, nil
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
	cli := c.auth.NewClient(token)

	opt := &spotify.PlayOptions{DeviceID: nil}
	if deviceID != "" {
		spotifyID := spotify.ID(deviceID)
		opt = &spotify.PlayOptions{DeviceID: &spotifyID}
	}

	err := cli.PlayOpt(opt)
	if convErr := c.convertPlayerError(err); convErr != nil {
		return fmt.Errorf("spotify api: play or resume: %w", convErr)
	}
	return nil
}

// Pause は再生を一時停止します。
// APIが非同期で処理がされるため、リクエストが返ってきても再生が一時停止されているとは限りません。
// 設定が反映されたか確認するには CurrentlyPlaying() を叩く必要があります。
// プレミアム会員必須
func (c *Client) Pause(ctx context.Context) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := c.auth.NewClient(token)

	// TODO : デバイスIDを指定する必要がある場合はいじる
	opt := &spotify.PlayOptions{DeviceID: nil}
	err := cli.PauseOpt(opt)
	if convErr := c.convertPlayerError(err); convErr != nil {
		return fmt.Errorf("spotify api: pause: %w", convErr)
	}
	return nil
}

// AddToQueue は曲を「次に再生される曲」に追加するAPIです。
// APIが非同期で処理がされるため、リクエストが返ってきても曲の追加が完了しているとは限りません。
// 設定が反映されたか確認するには CurrentlyPlaying() を叩く必要があります。
// プレミアム会員必須
func (c *Client) AddToQueue(ctx context.Context, trackID string) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := c.auth.NewClient(token)

	// TODO : デバイスIDを指定する必要がある場合はいじる
	opt := &spotify.PlayOptions{DeviceID: nil}
	err := cli.QueueSongOpt(spotify.ID(trackID), opt)
	if convErr := c.convertPlayerError(err); convErr != nil {
		return fmt.Errorf("spotify api: add queue: %w", convErr)
	}
	return nil
}

// SetRepeatMode はリピートモードの設定を変更するAPIです。
// APIが非同期で処理がされるため、リクエストが返ってきてもリピートモードの設定が完了しているとは限りません。
// 設定が反映されたか確認するには CurrentlyPlaying() を叩く必要があります。
// プレミアム会員必須
func (c *Client) SetRepeatMode(ctx context.Context, on bool) error {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return errors.New("token not found")
	}
	cli := c.auth.NewClient(token)

	state := "off"
	if on {
		state = "context"
	}

	// TODO : デバイスIDを指定する必要がある場合はいじる
	opt := &spotify.PlayOptions{DeviceID: nil}

	if err := cli.RepeatOpt(state, opt); c.convertPlayerError(err) != nil {
		return fmt.Errorf("spotify api: set repeat mode: %w", c.convertPlayerError(err))
	}
	return nil
}

func (c *Client) convertPlayerError(err error) error {
	if e, ok := err.(spotify.Error); ok {
		switch {
		case e.Status == http.StatusForbidden && strings.Contains(e.Message, "Restriction violated"):
			// https://github.com/spotify/web-api/issues/1205
			fmt.Printf("already in the ideal state: %s\n", e.Message)
			return nil
		case e.Status == http.StatusForbidden:
			return fmt.Errorf("%s: %w", e.Message, entity.ErrNonPremium)
		case e.Status == http.StatusNotFound:
			return fmt.Errorf("%s: %w", e.Message, entity.ErrActiveDeviceNotFound)
		}
	}
	return err
}
