package handler

import (
	"net/http"

	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"
	"github.com/labstack/echo/v5"
)

func (h *HandlerComplex) HandleHealthCheckDb(ctx *echo.Context) error {
	ok, err := h.ServiceHealth.HealthCheckDb()
	if err != nil {
		return err
	}
	if !ok {
		return errs.BuildErrConnDb(errs.ConnDb, http.StatusInternalServerError)
	}
	return mid.Respond(ctx, http.StatusOK, string(consts.HttpExprOk))
}
