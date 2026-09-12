package handler

import (
	"net/http"

	"github.com/atomreforge/daizy-night-server/internal/consts"
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"

	"github.com/labstack/echo/v5"
)

// HandleNotifGet is a placeholder of the notification feature (init stage).
func (h *HandlerComplex) HandleNotifGet(ctx *echo.Context) error {
	return mid.Respond(ctx, http.StatusOK, string(consts.HttpExprOk))
}

// HandleNotifPost is a placeholder of the notification feature (init stage).
func (h *HandlerComplex) HandleNotifPost(ctx *echo.Context) error {
	return mid.Respond(ctx, http.StatusOK, string(consts.HttpExprOk))
}
