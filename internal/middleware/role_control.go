package middleware

import (
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

// Roles guard an endpoint by requiring the authenticated user to hold
// one of the given roles. It must run AFTER AuthenJWT.
func RoleControl(allowed ...consts.Role) echo.MiddlewareFunc {
	allow := make(map[consts.Role]bool, len(allowed))
	for _, r := range allowed {
		allow[r] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token, err := echo.ContextGet[*jwt.Token](c, string(consts.CtxExprKeyJWT))
			if err != nil {
				return echo.ErrUnauthorized
			}
			claims, ok := token.Claims.(*model.JwtAccessTokenPayload)
			if !ok {
				return echo.ErrUnauthorized
			}
			if !allow[claims.Role] {
				// user is authenticated but not permitted
				return echo.ErrForbidden // 403
			}
			return next(c)
		}
	}
}
