package middleware

import (
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/labstack/echo/v5"
)

func MarkDomain(domain consts.RouterDomain) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) error {
			ctx.Set(string(consts.CtxExprKeyDomain), domain)
			return next(ctx)
		}
	}
}
