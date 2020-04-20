package handler

import (
	"github.com/labstack/echo/v4"
)

type TrackHandler struct {
}

func NewTrackHandler() *TrackHandler {
	return &TrackHandler{}
}

func (h *TrackHandler) SearchTracks(c echo.Context) error {
	return nil
}
