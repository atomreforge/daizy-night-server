package handler

import (
	"net/http"

	v1 "github.com/atomreforge/daizy-night-server/internal/api/v1/user"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"
	"github.com/atomreforge/daizy-night-server/internal/utils"
	"github.com/labstack/echo/v5"
)

func (h *HandlerComplex) HandleRefreshAccessToken(ctx *echo.Context) error {
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerRefresh))
	utils.Layer(ctx).Info(string(consts.ExprReqRefresh))

	req, err := Bind[v1.RefrehAccessTokenRequest](ctx)
	if err != nil {
		return err
	}

	// a missing token is 400 validation, not 401 lookup-miss
	if err := ValidateRefreshParams(req.RefreshToken); err != nil {
		return err
	}

	ok, accessToken, refreshToken, err := h.ServiceUser.RefreshAccessToken(req.RefreshToken)
	if !ok || err != nil {
		return err
	}
	return mid.RespondObj(ctx, http.StatusOK, v1.RefreshAccessTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
