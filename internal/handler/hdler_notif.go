package handler

import (
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"
	"github.com/labstack/echo/v5"
)

func (h *HandlerComplex) HandleNotifGet(ctx *echo.Context) error {
	return mid.Respond()
}

func (h *HandlerComplex) HandleNotifPost(ctx *echo.Context) error {
	return mid.Respond()
}
