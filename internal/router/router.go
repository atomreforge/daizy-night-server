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

	// owner user
	// user-scoped routes carry a :username segment; handlers must verify it
	// matches the authenticated identity (requireSelf), uid/ownership always
	// comes from the JWT claims.
	ptUser := e.Group("/api/v1/user")
	ptUser.Use(mid.AuthenJWT(pCrypto))
	ptUser.Use(mid.MarkDomain(consts.DomainPrivate))
	ptUser.POST("/signout", h.HandleSignout)
	ptUser.GET("/:username/info", h.HandleInfo)
	// services
	ptUser.GET("/:username/calendar", h.HandleCalendarGet)
	ptUser.PUT("/:username/calendar", h.HandleCalendarPut)
	ptUser.DELETE("/:username/calendar", h.HandleCalendarDelete)

	// any user
	ptAuthOnly := e.Group("/api/v1/public")
	ptAuthOnly.Use(mid.AuthenJWT(pCrypto))
	ptAuthOnly.Use(mid.MarkDomain(consts.DomainUnauthorized))
	ptAuthOnly.GET("/health/db", h.HandleHealthCheckDb)

	ptAuthOnly.GET("/notif/get/single", h.HandleNotificationGetSingle)
	ptAuthOnly.GET("/notif/get/batch", h.HandleNotificationGetBatch)
	ptAuthOnly.GET("/notif/get/update", h.HandleNotificationGetBatchUpdate)

	ptPublic := ptAuthOnly.Group("/user")
	// AuthenJWT/DomainUnauthorized are inherited from ptAuthOnly; override as public domain
	ptPublic.Use(mid.MarkDomain(consts.DomainPublic))
	ptPublic.GET("/:username/info", h.HandleInfo)
	ptPublic.GET("/:username/calendar", h.HandleCalendarGet)

	// admin only endpoints
	ptAdmin := e.Group("/api/v1/admin")
	ptAdmin.Use(mid.AuthenJWT(pCrypto))
	ptAdmin.Use(mid.RoleControl(consts.Admin))
	ptAdmin.POST("/sudo", h.HandleAdminSudo)
	ptAdmin.POST("/notif/post", h.HandleNotificationPost)
	ptAdmin.POST("/notif/update", h.HandleNotificationUpdate)
	ptAdmin.DELETE("/notif/delete", h.HandleNotificationDelete)
	ptAdmin.POST("/notif/withdraw", h.HandleNotificationWithdraw)

	return e
}
