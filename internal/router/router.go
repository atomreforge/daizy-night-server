package router

import (
	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/config"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/handler"
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

func New(
	h *handler.HandlerComplex,
	pCrypto abstract.InterfaceCrypto,
	cfg *config.Config,
) *echo.Echo {
	// init
	e := echo.New()

	// Recover is registered first so it also covers every middleware below
	e.Use(echomw.Recover())
	e.Use(mid.Inject())                  // injector
	e.Use(mid.RateLimit(cfg))            // rate limiter
	e.Use(echomw.BodyLimit(1024 * 1024)) // cap request bodies at 1 MiB, protects the db from abuse

	// public routes
	e.POST("/api/v1/register", h.HandleRegister)
	e.POST("/api/v1/login", h.HandleLogin)
	e.POST("/api/v1/refresh-access-token", h.HandleRefreshAccessToken)

	// endpoints that requires authen only.
	// user-scoped routes carry a :username segment; handlers must verify it
	// matches the authenticated identity (requireSelf), uid/ownership always
	// comes from the JWT claims.
	ptAuthOnly := e.Group("/api/v1")
	ptAuthOnly.Use(mid.AuthenJWT(pCrypto))
	ptAuthOnly.GET("/user/:username/me", h.HandleMe)
	ptAuthOnly.POST("/user/signout", h.HandleSignout)
	ptAuthOnly.GET("/user/:username/calendar", h.HandleCalendarGet)
	ptAuthOnly.PUT("/user/:username/calendar", h.HandleCalendarPut)
	ptAuthOnly.DELETE("/user/:username/calendar", h.HandleCalendarDelete)
	ptAuthOnly.GET("/health/db", h.HandleHealthCheckDb)
	ptAuthOnly.GET("/notif/get", h.HandleNotifGet())

	// admin only endpoints
	ptAdmin := e.Group("/api/v1/admin")
	ptAdmin.Use(mid.AuthenJWT(pCrypto))
	ptAdmin.Use(mid.RoleControl(consts.Admin))
	ptAdmin.POST("/sudo", h.HandleAdminSudo)
	ptAdmin.POST("/notif/post", h.HandleNotifPost())

	return e
}
