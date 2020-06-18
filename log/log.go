package log

import (
	"github.com/camphor-/relaym-server/config"
	"github.com/labstack/gommon/log"
)

// New はロガーを生成します。
func New() *log.Logger {
	logger := log.New("application")
	logger.SetLevel(log.INFO)
	if config.IsLocal() || config.IsDev() {
		logger.SetLevel(log.DEBUG)
	}
	return logger
}
