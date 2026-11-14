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

func (h *HandlerComplex) HandleLogin(ctx *echo.Context) error {
	// record flow chain (monotonically accumulating)
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerLogin))

	req, err := Bind[v1.LoginRequest](ctx)
	if err != nil {
		return err
	}

	utils.Layer(ctx).Info(fmt.Sprintf("%s; user::%s", consts.ExprReqLogin, req.Username))

	if err := ValidateLoginParams(&req.LoginBody); err != nil {
		return err
	}

	ok, accessToken, refreshToken, err := h.ServiceUser.Login(&req.LoginBody)
	utils.AppendCallChain(ctx, string(consts.ModExprServiceUser)) // call chain
	if !ok {
		return err
	}

	utils.Layer(ctx).Info(fmt.Sprintf("successfully logged in user::%s", req.Username))
	return mid.RespondObj(ctx, http.StatusOK, v1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
