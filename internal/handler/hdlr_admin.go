package handler

import (
	"fmt"

	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/utils"

	"github.com/labstack/echo/v5"
)

func (h *HandlerComplex) HandleAdminSudo(ctx *echo.Context) error {
	/* common chore*/
	// record flow chain (monotonically accumulating)
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerAdmin))
	utils.AppendCallChain(ctx, string(consts.ModExprServiceAdmin))
	utils.Layer(ctx).Info(fmt.Sprintf("%s;", consts.ExprReqAdminSudo))

	h.ServiceAdmin.Sudo()
	return nil
}
