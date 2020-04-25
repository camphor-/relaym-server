package handler

import (
	"net/http"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/labstack/echo/v4"
)

type TrackHandler struct {
	trackUC *usecase.TrackUseCase
}

func NewTrackHandler(trackUC *usecase.TrackUseCase) *TrackHandler {
	return &TrackHandler{trackUC: trackUC}
}

func (h *TrackHandler) SearchTracks(c echo.Context) error {
	q := c.QueryParam("q")
	ctx := c.Request().Context()
	tracks, err := h.trackUC.SearckTracks(ctx, q)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "query is empty")
	}
	return c.JSON(http.StatusOK, &tracksRes{
		Tracks: toTrackJSON(tracks),
	})
}

func toTrackJSON(tracks []*entity.Track) []*trackJSON {
	trackJSONArray := make([]*trackJSON, len(tracks))

	for i, track := range tracks {
		trackJSONArray[i] = &trackJSON{
			URI:      track.URI,
			ID:       track.ID,
			Name:     track.Name,
			Duration: track.Duration,
			Artists:  toArtistJSON(track),
			URL:      track.URL,
			Album: &albumJSON{
				Name:   track.Album.Name,
				Images: toAlbumImageJSON(track),
			},
		}
	}

	return trackJSONArray
}

func toArtistJSON(track *entity.Track) []*artistJSON {
	artistJSONArray := make([]*artistJSON, len(track.Artists))
	for i, artist := range track.Artists {
		artistJSONArray[i] = &artistJSON{
			Name: artist.Name,
		}
	}

	return artistJSONArray
}

func toAlbumImageJSON(track *entity.Track) []*albumImageJSON {
	albumImageJSONArray := make([]*albumImageJSON, len(track.Album.Images))
	for i, albumImage := range track.Album.Images {
		albumImageJSONArray[i] = &albumImageJSON{
			URL:    albumImage.URL,
			Height: albumImage.Height,
			Width:  albumImage.Width,
		}
	}

	return albumImageJSONArray
}

type tracksRes struct {
	Tracks []*trackJSON `json:"tracks"`
}

type trackJSON struct {
	URI      string        `json:"uri"`
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration_ms"`
	Artists  []*artistJSON `json:"artists"`
	URL      string        `json:"external_urls"`
	Album    *albumJSON    `json:"album"`
}

type albumJSON struct {
	Name   string            `json:"name"`
	Images []*albumImageJSON `json:"images"`
}

type albumImageJSON struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type artistJSON struct {
	Name string `json:"name"`
}
