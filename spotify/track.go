package spotify

import (
	"context"
	"fmt"
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

func (c *Client) toTracks(resultTracks []spotify.FullTrack) []*entity.Track {
	tracks := make([]*entity.Track, len(resultTracks))

	for i, rt := range resultTracks {
		tracks[i] = &entity.Track{
			URI:      string(rt.URI),
			ID:       rt.ID.String(),
			Name:     rt.Name,
			Duration: time.Duration(rt.Duration) * time.Millisecond,
			Artists:  c.toArtists(rt.Artists),
			URL:      rt.ExternalURLs["spotify"],
			Album: &entity.Album{
				Name:   rt.Album.Name,
				Images: c.toImages(rt.Album.Images),
			},
		}
	}

	return tracks
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
