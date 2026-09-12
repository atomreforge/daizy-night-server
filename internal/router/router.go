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

	// anonymous endpoints
	e.POST("/api/v1/register", h.HandleRegister)
	e.POST("/api/v1/login", h.HandleLogin)
	e.POST("/api/v1/refresh-access-token", h.HandleRefreshAccessToken)

	// endpoints that requires authen only.
	// user-scoped routes carry a :username segment; handlers must verify it
	// matches the authenticated identity (requireSelf), uid/ownership always
	// comes from the JWT claims.
	ptUser := e.Group("/api/v1/user")
	ptUser.Use(mid.AuthenJWT(pCrypto))
	ptUser.Use(mid.MarkDomain(consts.DomainPrivate))
	// session manage
	ptUser.POST("/signout", h.HandleSignout)
	// user info
	ptUser.GET("/:username/info", h.HandleInfo)
	// services
	ptUser.GET("/:username/calendar", h.HandleCalendarGet)
	ptUser.PUT("/:username/calendar", h.HandleCalendarPut)
	ptUser.DELETE("/:username/calendar", h.HandleCalendarDelete)

	// public domain; any verified user;
	ptAuthOnly := e.Group("/api/v1/public")
	ptAuthOnly.GET("/health/db", h.HandleHealthCheckDb)
	ptAuthOnly.GET("/notif/get", h.HandleNotifGet)

	ptPublic := ptAuthOnly.Group("/user")
	ptPublic.Use(mid.AuthenJWT(pCrypto))
	ptPublic.Use(mid.MarkDomain(consts.DomainPublic))
	ptPublic.GET("/:username/info", h.HandleInfo)
	ptPublic.GET("/:username/calendar", h.HandleCalendarGet)

	// admin only endpoints
	ptAdmin := e.Group("/api/v1/admin")
	ptAdmin.Use(mid.AuthenJWT(pCrypto))
	ptAdmin.Use(mid.RoleControl(consts.Admin))
	ptAdmin.POST("/sudo", h.HandleAdminSudo)
	ptAdmin.POST("/notif/post", h.HandleNotifPost)

	return e
}
