package handler

import (
	"fmt"
	"net/http"

	v1 "github.com/atomreforge/daizy-night-server/internal/api/v1"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"
	"github.com/atomreforge/daizy-night-server/internal/utils"

	"github.com/labstack/echo/v5"
)

func (h *HandlerComplex) HandleRegister(ctx *echo.Context) error {
	// record flow chain (monotonically accumulating)
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerRegister))

	req, err := Bind[v1.RegisterRequest](ctx)
	if err != nil {
		return err
	}

	utils.Layer(ctx).Info(fmt.Sprintf("%s; user::%s;", consts.ExprReqRegister, req.Username))

	// failure during param validation
	if err := ValidateRegisterParams(req.RegisterBody); err != nil {
		return err
	}

	// execute reg service
	utils.AppendCallChain(ctx, string(consts.ModExprServiceUser))
	user, err := h.ServiceUser.Register(&req.RegisterBody)
	if err != nil {
		//h.ServiceCode.RemoveRegistercode(req.Registercode)
		return err
	}

	utils.Layer(ctx).Info(fmt.Sprintf("successfully registered; user::%s, uid::%d", user.Username, user.ID))
	return mid.Respond(ctx, http.StatusOK, string(consts.HttpExprOk))
}
