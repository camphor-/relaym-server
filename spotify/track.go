package spotify

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/zmb3/spotify"
)

// Search はSpotify APIを通して、与えられたクエリを用い音楽を検索します。
func (c *Client) Search(ctx context.Context, q string) ([]*entity.Track, error) {
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("token not found")
	}

	cli := c.auth.NewClient(token)
	result, err := cli.Search(q, spotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("search q=%s: %w", q, err)
	}
	return c.toTracks(result.Tracks.Tracks), nil
}

// GetTrackFromURI はSpotify APIを通して、与えられたTrack URIを用い音楽を取得します。
func (c *Client) GetTracksFromURI(ctx context.Context, trackURIs []string) ([]*entity.Track, error) {
	if len(trackURIs) == 0 {
		return nil, nil
	}
	token, ok := service.GetTokenFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("token not found")
	}
	cli := c.auth.NewClient(token)

	ids := make([]spotify.ID, len(trackURIs))
	for i, trackURI := range trackURIs {
		id := strings.Replace(trackURI, "spotify:track:", "", 1)
		ids[i] = spotify.ID(id)
	}

	tracks := make([]*entity.Track, len(trackURIs))

	// GetTracksは一度につき50曲までしか取得できない
	countForLoop := int(math.Ceil(float64(len(ids)) / 50.0))
	for i := 0; i < countForLoop; i++ {
		idsForAPI := ids[countForLoop*50 : (countForLoop+1)*50]
		resultTracks, err := cli.GetTracks(idsForAPI...)
		if err != nil {
			return nil, fmt.Errorf("get track uris=%s: %w", trackURIs, err)
		}

		for j, rt := range resultTracks {
			idx := i + j
			tracks[idx] = c.toTrack(rt)
		}

	}

	return tracks, nil
}

func (c *Client) toTracks(resultTracks []spotify.FullTrack) []*entity.Track {
	tracks := make([]*entity.Track, len(resultTracks))

	for i, rt := range resultTracks {
		rt := rt
		tracks[i] = c.toTrack(&rt)
	}

	return tracks
}

func (c *Client) toTrack(fullTrack *spotify.FullTrack) *entity.Track {
	if fullTrack == nil {
		return nil
	}
	return &entity.Track{
		URI:      string(fullTrack.URI),
		ID:       fullTrack.ID.String(),
		Name:     fullTrack.Name,
		Duration: time.Duration(fullTrack.Duration) * time.Millisecond,
		Artists:  c.toArtists(fullTrack.Artists),
		URL:      fullTrack.ExternalURLs["spotify"],
		Album: &entity.Album{
			Name:   fullTrack.Album.Name,
			Images: c.toImages(fullTrack.Album.Images),
		},
	}
}

func (c *Client) toArtists(resultArtists []spotify.SimpleArtist) []*entity.Artist {
	artists := make([]*entity.Artist, len(resultArtists))
	for i, a := range resultArtists {
		artists[i] = &entity.Artist{
			Name: a.Name,
		}
	}
	return artists
}

func (c *Client) toImages(resultImages []spotify.Image) []*entity.AlbumImage {
	images := make([]*entity.AlbumImage, len(resultImages))

	for i, image := range resultImages {
		images[i] = &entity.AlbumImage{
			URL:    image.URL,
			Height: image.Height,
			Width:  image.Width,
		}
	}
	return images
}
