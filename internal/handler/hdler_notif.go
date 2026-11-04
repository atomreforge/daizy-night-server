package handler

import (
	"net/http"

	"github.com/atomreforge/daizy-night-server/internal/consts"
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"
	"github.com/labstack/echo/v5"
)

func (h *HandlerComplex) HandleNotifGet(ctx *echo.Context) error {
	return mid.Respond(ctx, http.StatusInternalServerError, string(consts.ExprUnsupportedFeature))
}

func (h *HandlerComplex) HandleNotifPost(ctx *echo.Context) error {
	return mid.Respond(ctx, http.StatusInternalServerError, string(consts.ExprUnsupportedFeature))
}
