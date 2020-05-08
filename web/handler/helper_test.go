package handler

import (
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func setToContext(c echo.Context, userID string, token *oauth2.Token) echo.Context {
	ctx := c.Request().Context()
	ctx = service.SetUserIDToContext(ctx, userID)
	ctx = service.SetTokenToContext(ctx, token)
	c.SetRequest(c.Request().WithContext(ctx))
	return c
}
