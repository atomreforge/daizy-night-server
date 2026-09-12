package handler

import (
	"fmt"
	"net/http"

	v1 "github.com/atomreforge/daizy-night-server/internal/api/v1/user"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

func (h *HandlerComplex) HandleInfo(ctx *echo.Context) error {
	// record flow chain (monotonically accumulating)
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerMe))
	utils.Layer(ctx).Info(fmt.Sprintf("%s", consts.ExprReqUserInfo))

	// public /info disclosure semantics (email, github ids) are not
	// settled yet; reject until a sanitized response is defined.
	if GetDomain(ctx) == consts.DomainPublic {
		return errs.BuildErrSupport(errs.FeatureUnsupported, http.StatusForbidden)
	}

	token, err := echo.ContextGet[*jwt.Token](ctx, string(consts.CtxExprKeyJWT))
	if err != nil {
		return err
	}
	claims, ok := token.Claims.(*model.JwtAccessTokenPayload)
	if !ok {
		utils.Layer(ctx).Error("failed to assert claims to JwtAccessTokenPayload")
		return echo.ErrUnauthorized
	}

	// the :username path segment must denote the authenticated user
	if err := requireSelf(ctx, claims); err != nil {
		return err
	}

	b, err := h.ServiceUser.GetInfoMineByUid(claims.Uid)
	utils.AppendCallChain(ctx, string(consts.ModExprServiceUser))
	if err != nil {
		return err
	}

	mid.RespondObj(ctx, http.StatusOK, v1.InfoMeResponse{InfoUser: *b})

	return nil
}
