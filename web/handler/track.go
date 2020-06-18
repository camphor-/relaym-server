package handler

import (
	"net/http"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/log"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/labstack/echo/v4"
)

// TrackHandler は/search を管理する構造体です。
type TrackHandler struct {
	trackUC *usecase.TrackUseCase
}

// NewTrackHandler はTrackHandlerのポインタを生成する関数です。
func NewTrackHandler(trackUC *usecase.TrackUseCase) *TrackHandler {
	return &TrackHandler{trackUC: trackUC}
}

// SearchTracks は GET /search に対応するハンドラーです。
func (h *TrackHandler) SearchTracks(c echo.Context) error {
	logger := log.New()
	q := c.QueryParam("q")
	if q == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query is empty")
	}

	ctx := c.Request().Context()
	tracks, err := h.trackUC.SearckTracks(ctx, q)
	if err != nil {
		logger.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, &tracksRes{
		Tracks: toTrackJSON(tracks),
	})
}

func toTrackJSON(tracks []*entity.Track) []*trackJSON {
	trackJSONs := make([]*trackJSON, len(tracks))

	for i, track := range tracks {
		trackJSONs[i] = &trackJSON{
			URI:      track.URI,
			ID:       track.ID,
			Name:     track.Name,
			Duration: track.Duration.Milliseconds(),
			Artists:  toArtistJSON(track),
			URL:      track.URL,
			Album: &albumJSON{
				Name:   track.Album.Name,
				Images: toAlbumImageJSON(track),
			},
		}
	}

	return trackJSONs
}

func toArtistJSON(track *entity.Track) []*artistJSON {
	artistJSONs := make([]*artistJSON, len(track.Artists))
	for i, artist := range track.Artists {
		artistJSONs[i] = &artistJSON{
			Name: artist.Name,
		}
	}

	return artistJSONs
}

func toAlbumImageJSON(track *entity.Track) []*albumImageJSON {
	albumImageJSONs := make([]*albumImageJSON, len(track.Album.Images))
	for i, albumImage := range track.Album.Images {
		albumImageJSONs[i] = &albumImageJSON{
			URL:    albumImage.URL,
			Height: albumImage.Height,
			Width:  albumImage.Width,
		}
	}

	return albumImageJSONs
}

type tracksRes struct {
	Tracks []*trackJSON `json:"tracks"`
}

type trackJSON struct {
	URI      string        `json:"uri"`
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Duration int64         `json:"duration_ms"`
	Artists  []*artistJSON `json:"artists"`
	URL      string        `json:"external_url"`
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
