package handler

import (
	"errors"
	"net/http"

	"github.com/camphor-/relaym-server/log"

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
	logger := log.New()
	type reqJSON struct {
		Name string `json:"name"`
	}
	req := new(reqJSON)
	if err := c.Bind(req); err != nil {
		logger.Errorj(map[string]interface{}{"message": "failed to bind", "error": err.Error()})
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
		logger.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, h.toSessionRes(session, nil, nil))
}

// GetSession は GET /sessions/:id に対応するハンドラーです。
func (h *SessionHandler) GetSession(c echo.Context) error {
	logger := log.New()
	ctx := c.Request().Context()
	id := c.Param("id")

	session, tracks, playingInfo, err := h.uc.GetSession(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrSessionNotFound) {
			logger.Debug(err)
			return echo.NewHTTPError(http.StatusNotFound)
		}
		logger.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, h.toSessionRes(session, playingInfo.Device, tracks))
}

// AddQueue は POST /sessions/:id/queue に対応するハンドラーです。
func (h *SessionHandler) AddQueue(c echo.Context) error {
	logger := log.New()
	type reqJSON struct {
		URI string `json:"uri"`
	}
	req := new(reqJSON)
	if err := c.Bind(req); err != nil {
		logger.Debug(err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid track id")
	}

	if req.URI == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid track id")
	}

	ctx := c.Request().Context()
	sessionID := c.Param("id")

	if err := h.uc.AddQueueTrack(ctx, sessionID, req.URI); err != nil {
		if errors.Is(err, entity.ErrSessionNotFound) {
			logger.Debug(err)
			return echo.NewHTTPError(http.StatusNotFound, entity.ErrSessionNotFound.Error())
		}
		logger.Errorj(map[string]interface{}{"message": "add queue track", "error": err.Error()})
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// Playback は PUT /sessions/:id/playback に対応するハンドラーです。
func (h *SessionHandler) Playback(c echo.Context) error {
	logger := log.New()
	type reqJSON struct {
		State string `json:"state"`
	}
	req := new(reqJSON)
	if err := c.Bind(req); err != nil {
		logger.Debugj(map[string]interface{}{"message": "failed to bind", "error": err.Error()})
		return echo.NewHTTPError(http.StatusBadRequest, "invalid state")
	}

	st, err := entity.NewStateType(req.State)
	if err != nil {
		logger.Debugj(map[string]interface{}{"message": "failed to convert state type", "error": err.Error()})
		return echo.NewHTTPError(http.StatusBadRequest, "invalid state")
	}

	ctx := c.Request().Context()
	sessionID := c.Param("id")
	if err := h.uc.ChangePlaybackState(ctx, sessionID, st); err != nil {
		switch {
		case errors.Is(err, entity.ErrSessionNotFound):
			logger.Debug(err)
			return echo.NewHTTPError(http.StatusNotFound, entity.ErrSessionNotFound.Error())
		case errors.Is(err, entity.ErrActiveDeviceNotFound):
			logger.Debug(err)
			return echo.NewHTTPError(http.StatusForbidden, entity.ErrActiveDeviceNotFound.Error())
		}
		logger.Errorj(map[string]interface{}{"message": "failed to change playback", "error": err.Error()})
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusAccepted)
}

// SetDevice PUT /sessions/:id/devicesに対応するハンドラーです。
func (h *SessionHandler) SetDevice(c echo.Context) error {
	logger := log.New()
	type reqJSON struct {
		DeviceID string `json:"device_id"`
	}
	req := new(reqJSON)
	if err := c.Bind(req); err != nil {
		logger.Debugj(map[string]interface{}{"message": "failed to bind", "error": err.Error()})
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
			logger.Debug(err)
			return echo.NewHTTPError(http.StatusNotFound, entity.ErrSessionNotFound.Error())
		case errors.Is(err, entity.ErrUserIsNotSessionCreator):
			logger.Debug(err)
			return echo.NewHTTPError(http.StatusForbidden, entity.ErrUserIsNotSessionCreator.Error())

		}
		logger.Errorj(map[string]interface{}{"message": "failed to set device", "error": err.Error(), "deviceID": req.DeviceID})
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *SessionHandler) toSessionRes(session *entity.SessionWithUser, device *entity.Device, tracks []*entity.Track) *sessionRes {
	var dJ *deviceJSON = nil
	if device != nil {
		dJ = &deviceJSON{
			ID:           device.ID,
			IsRestricted: device.IsRestricted,
			Name:         device.Name,
		}
	}
	return &sessionRes{
		ID:   session.ID,
		Name: session.Name,
		Creator: creatorJSON{
			ID:          session.Creator.ID,
			DisplayName: session.Creator.DisplayName,
		},
		Playback: playbackJSON{
			State: stateJSON{
				Type: session.StateType.String(),
			},
			Device: dJ,
		},
		Queue: queueJSON{
			Head:   session.QueueHead,
			Tracks: toTrackJSON(tracks),
		},
	}
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
	State  stateJSON   `json:"state"`
	Device *deviceJSON `json:"device"`
}
type stateJSON struct {
	Type string `json:"type"`
}

type queueJSON struct {
	Head   int          `json:"head"`
	Tracks []*trackJSON `json:"tracks"`
}
