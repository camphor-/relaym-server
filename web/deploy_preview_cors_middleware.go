package web

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/camphor-/relaym-server/config"

	"github.com/labstack/echo/v4"
)

type deployPreviewCorsMiddleware struct {
	re               *regexp.Regexp
	allowHeaders     []string
	allowCredentials bool
}

func newDeployPreviewCorsMiddleware(allowHeaders []string, allowCredentials bool) *deployPreviewCorsMiddleware {
	return &deployPreviewCorsMiddleware{
		re:               regexp.MustCompile("https://deploy-preview-[0-9]+--relaym.netlify.app"),
		allowHeaders:     allowHeaders,
		allowCredentials: allowCredentials,
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
		return next(c)
	}
}

func (m *deployPreviewCorsMiddleware) addAllowOriginForOption(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()

		if req.Method != http.MethodOptions {
			return next(c)
		}

		origin := req.Header.Get(echo.HeaderOrigin)

		allowMethods := strings.Join([]string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete}, ",")

		// Preflight request
		res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)
		res.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestMethod)
		res.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestHeaders)
		res.Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
		res.Header().Set(echo.HeaderAccessControlAllowMethods, allowMethods)
		if m.allowCredentials {
			res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
		}
		if allowHeaders := strings.Join(m.allowHeaders, ","); allowHeaders != "" {
			res.Header().Set(echo.HeaderAccessControlAllowHeaders, allowHeaders)
		} else {
			h := req.Header.Get(echo.HeaderAccessControlRequestHeaders)
			if h != "" {
				res.Header().Set(echo.HeaderAccessControlAllowHeaders, h)
			}
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func (m *deployPreviewCorsMiddleware) IsDeployPreviewOrigin(origin string) bool {
	return m.re.MatchString(origin)
}
