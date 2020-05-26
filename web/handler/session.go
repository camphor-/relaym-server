package handler

import (
	"errors"
	"net/http"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/service"
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

// PostSession は POST /sessions に対応するハンドラーです。
func (h *SessionHandler) PostSession(c echo.Context) error {
	type reqJSON struct {
		Name string `json:"name"`
	}
	req := new(reqJSON)
	if err := c.Bind(req); err != nil {
		c.Logger().Debugf("bind: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	sessionName := req.Name
	if sessionName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty name")
	}

	ctx := c.Request().Context()
	userID, _ := service.GetUserIDFromContext(ctx)
	session, err := h.uc.CreateSession(sessionName, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, h.toSessionRes(session))
}

func (h *SessionHandler) toSessionRes(session *entity.SessionWithUser) *sessionRes {
	return &sessionRes{
		ID:   session.ID,
		Name: session.Name,
		Creator: creatorJSON{
			ID:          session.Creator.ID,
			DisplayName: session.Creator.DisplayName,
		},
		Playback: playbackJSON{
			State: stateJSON{
				Type: "STOP",
			},
			Device: nil, //TODO: deviceを取得し、deviceJSONを作成する
		},
		Queue: queueJSON{
			Head:   session.QueueHead,
			Tracks: nil, //TODO: queueTrackのsessionIDからsessionを取得し、trackJSONを作成する
		},
	}
}

// AddQueue は POST /sessions/:id/queue に対応するハンドラーです。
func (h *SessionHandler) AddQueue(c echo.Context) error {
	type reqJSON struct {
		State string `json:"uri"`
	}
	req := new(reqJSON)
	if err := c.Bind(req); err != nil {
		c.Logger().Debugf("bind: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid track id")
	}

	ctx := c.Request().Context()
	sessionID := c.Param("id")

	if err := h.uc.AddQueueTrack(ctx, sessionID, req.State); err != nil {
		if errors.Is(err, entity.ErrSessionNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, entity.ErrSessionNotFound.Error())
		}
		c.Logger().Errorf("add queue track: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusAccepted)
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

// SetDevice PUT /sessions/:id/devicesに対応するハンドラーです。
func (h *SessionHandler) SetDevice(c echo.Context) error {
	type reqJSON struct {
		DeviceID string `json:"device_id"`
	}
	req := new(reqJSON)
	if err := c.Bind(req); err != nil {
		c.Logger().Debugf("bind: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "empty device id")
	}

	if req.DeviceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty device id")
	}

	ctx := c.Request().Context()
	sessionID := c.Param("id")

	if err := h.uc.SetDevice(ctx, sessionID, req.DeviceID); err != nil {
		switch {
		case errors.Is(err, entity.ErrSessionNotFound):
			return echo.NewHTTPError(http.StatusNotFound, entity.ErrSessionNotFound.Error())
		case errors.Is(err, entity.ErrUserIsNotSessionCreator):
			return echo.NewHTTPError(http.StatusForbidden, entity.ErrUserIsNotSessionCreator.Error())

		}
		c.Logger().Errorf("set device id=%s: %v", req.DeviceID, err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

type sessionRes struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Creator  creatorJSON  `json:"creator"`
	Playback playbackJSON `json:"playback"`
	Queue    queueJSON    `json:"queue"`
}

type creatorJSON struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type playbackJSON struct {
	State  stateJSON     `json:"state"`
	Device []*deviceJSON `json:"device"`
}
type stateJSON struct {
	Type string `json:"type"`
}

type queueJSON struct {
	Head   int          `json:"head"`
	Tracks []*trackJSON `json:"tracks"`
}
