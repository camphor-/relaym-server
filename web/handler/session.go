package handler

import (
	"errors"
	"net/http"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/labstack/echo/v4"
)

// SessionHandler は /sessions 以下のエンドポイントを管理する構造体です。
type SessionHandler struct {
	uc *usecase.SessionUseCase
}

// NewSessionHandler はSessionHandlerのポインタを生成する関数です。
func NewSessionHandler(uc *usecase.SessionUseCase) *SessionHandler {
	return &SessionHandler{uc: uc}
}

// Playback は PUT /sessions/:id/playback に対応するハンドラーです。
func (h *SessionHandler) Playback(c echo.Context) error {
	type reqJSON struct {
		State string `json:"state"`
	}
	req := new(reqJSON)
	if err := c.Bind(req); err != nil {
		c.Logger().Debugf("bind: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid state")
	}

	st, err := entity.NewStateType(req.State)
	if err != nil {
		c.Logger().Debugf("NewStateType: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid state")
	}

	ctx := c.Request().Context()
	sessionID := c.Param("id")
	if err := h.uc.ChangePlaybackState(ctx, sessionID, st); err != nil {
		switch {
		case errors.Is(err, entity.ErrSessionNotFound):
			return echo.NewHTTPError(http.StatusNotFound, entity.ErrSessionNotFound.Error())
		case errors.Is(err, entity.ErrActiveDeviceNotFound):
			return echo.NewHTTPError(http.StatusForbidden, entity.ErrActiveDeviceNotFound.Error())
		}
		c.Logger().Errorf("change playback: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusAccepted)
}
