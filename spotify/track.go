package spotify

import (
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/zmb3/spotify"
)

func (c *Client) Search(q string) ([]*entity.Track, error) {
	cli := c.auth.NewClient(c.token)
	result, err := cli.Search(q, spotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("serach q=%s: %w", q, err)
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
