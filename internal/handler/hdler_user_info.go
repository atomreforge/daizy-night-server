package handler

import (
	"fmt"
	"net/http"

	v1public "github.com/atomreforge/daizy-night-server/internal/api/v1/public/user"
	v1 "github.com/atomreforge/daizy-night-server/internal/api/v1/user"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

func (h *HandlerComplex) HandleInfo(ctx *echo.Context) error {
	// record flow chain (monotonically accumulating)
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerMe))

	d := GetDomain(ctx)
	utils.Layer(ctx).Info(fmt.Sprintf(
		"%s; user::%s; domain::%s",
		consts.ExprReqUserInfo,
		ctx.Param("username"),
		d.Say(),
	))

	switch d {
	// public domain: sanitized view of any user's info — account-confidential
	// fields (email, github identity) are deliberately omitted; no ownership
	// check, same semantics as the public calendar endpoint.
	case consts.DomainPublic:
		info, err := h.ServiceUser.GetInfoByUsername(ctx.Param("username"))
		utils.AppendCallChain(ctx, string(consts.ModExprServiceUser))
		if err != nil {
			return err
		}

		return mid.RespondObj(ctx, http.StatusOK, v1public.ResponseAnyInfoGet{
			UserID:       info.Uid,
			Username:     info.Username,
			Nickname:     info.Nickname,
			RegisterTime: info.RegisterTime,
			Role:         info.Role,
		})

	// private domain: the full self view; the :username path segment must
	// denote the authenticated user (requireSelf), uid always comes from
	// the JWT claims.
	case consts.DomainPrivate:
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

		return mid.RespondObj(ctx, http.StatusOK, v1.InfoMeResponse{InfoUser: *b})
	}

	return mid.Respond(ctx, http.StatusBadRequest, string(consts.ExprUnreachableCase))
}
