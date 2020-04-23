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
	tracks, err := h.trackUC.SearckTracks(q)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "query is empty")
	}
	return c.JSON(http.StatusOK, &tracksRes{
		Tracks: toTrackJson(tracks),
	})
}

func toTrackJson(tracks []*entity.Track) []*trackJson {
	trackJsonArray := make([]*trackJson, len(tracks))

	for i, track := range tracks {
		trackJsonArray[i] = &trackJson{
			URI:      track.URI,
			ID:       track.ID,
			Name:     track.Name,
			Duration: track.Duration,
			Artists:  toArtistJson(track),
			URL:      track.URL,
			Album: &albumJson{
				Name:   track.Album.Name,
				Images: toAlbumImageJson(track),
			},
		}
	}

	return trackJsonArray
}

func toArtistJson(track *entity.Track) []*artistJson {
	artistJsonArray := make([]*artistJson, len(track.Artists))
	for i, artist := range track.Artists {
		artistJsonArray[i] = &artistJson{
			Name: artist.Name,
		}
	}

	return artistJsonArray
}

func toAlbumImageJson(track *entity.Track) []*albumImageJson {
	albumImageJsonArray := make([]*albumImageJson, len(track.Album.Images))
	for i, albumImage := range track.Album.Images {
		albumImageJsonArray[i] = &albumImageJson{
			URL:    albumImage.URL,
			Height: albumImage.Height,
			Width:  albumImage.Width,
		}
	}

	return albumImageJsonArray
}

type tracksRes struct {
	Tracks []*trackJson `json:"tracks"`
}

type trackJson struct {
	URI      string        `json:"uri"`
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration_ms"`
	Artists  []*artistJson `json:"artists"`
	URL      string        `json:"external_urls"`
	Album    *albumJson    `json:"album"`
}

type albumJson struct {
	Name   string            `json:"name"`
	Images []*albumImageJson `json:"images"`
}

type albumImageJson struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type artistJson struct {
	Name string `json:"name"`
}
