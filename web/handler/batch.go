package handler

import (
	"net/http"

	"github.com/camphor-/relaym-server/log"
	"github.com/camphor-/relaym-server/usecase"
	"github.com/labstack/echo/v4"
)

// BatchHandler は /batch 以下のエンドポイントを管理する構造体です。
type BatchHandler struct {
	uc *usecase.BatchUseCase
}

// NewBatchHandler はBatchHandlerのポインタを生成する関数です。
func NewBatchHandler(uc *usecase.BatchUseCase) *BatchHandler {
	return &BatchHandler{uc: uc}
}

// PostArchive は POST /archive に対応するハンドラーです。
func (h *BatchHandler) PostArchive(c echo.Context) error {
	logger := log.New()
	if err := h.uc.ArchiveOldSessions(); err != nil {
		logger.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}
