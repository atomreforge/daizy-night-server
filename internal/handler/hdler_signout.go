package handler

import (
	"fmt"
	"net/http"

	"github.com/atomreforge/daizy-night-server/internal/consts"
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

func (h *HandlerComplex) HandleSignout(ctx *echo.Context) error {
	// record flow chain (monotonically accumulating)
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerSignout))

	req, err := Bind[model.SignoutBody](ctx)
	if err != nil {
		return err
	}

	utils.Layer(ctx).Info(fmt.Sprintf("%s;", consts.ExprReqSignout))

	// failure during param validation
	if err := ValidateSignoutParams(&req); err != nil {
		return err
	}

	// the session belongs to the authenticated caller: uid never comes
	// from the request body
	token, err := echo.ContextGet[*jwt.Token](ctx, string(consts.CtxExprKeyJWT))
	if err != nil {
		return err
	}
	claims, ok := token.Claims.(*model.JwtAccessTokenPayload)
	if !ok {
		utils.Layer(ctx).Error("failed to assert claims to JwtAccessTokenPayload")
		return echo.ErrUnauthorized
	}

	// execute signout service; the uid is injected from the claims, never
	// trusted from the request payload
	utils.AppendCallChain(ctx, string(consts.ModExprServiceUser))
	req.Uid = claims.Uid
	if _, err := h.ServiceUser.Signout(&req); err != nil {
		return err
	}

	utils.Layer(ctx).Info(fmt.Sprintf("successfully signed out; user::%s, uid::%d", claims.Username, claims.Uid))
	return mid.Respond(ctx, http.StatusOK, string(consts.HttpExprOk))
}
