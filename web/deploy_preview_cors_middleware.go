package web

import (
	"net/http"
	"regexp"

	"github.com/camphor-/relaym-server/config"

	"github.com/labstack/echo/v4"
)

type deployPreviewCorsMiddleware struct {
	re *regexp.Regexp
}

func newDeployPreviewCorsMiddleware() *deployPreviewCorsMiddleware {
	return &deployPreviewCorsMiddleware{
		re: regexp.MustCompile("https://deploy-preview-[0-9]+--relaym.netlify.app"),
	}
}

func (m *deployPreviewCorsMiddleware) addAllowOrigin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()
		origin := req.Header.Get(echo.HeaderOrigin)

		if origin == config.CORSAllowOrigin() {
			return next(c)
		}

		if !m.IsDeployPreviewOrigin(origin) {
			return next(c)
		}

		if req.Method != http.MethodOptions {
			res.Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
			return next(c)
		}

		res.Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
		return c.NoContent(http.StatusNoContent)
	}
}

func (m *deployPreviewCorsMiddleware) IsDeployPreviewOrigin(origin string) bool {
	return m.re.MatchString(origin)
}
