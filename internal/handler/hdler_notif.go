package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	mid "github.com/atomreforge/daizy-night-server/internal/middleware"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"
	"github.com/labstack/echo/v5"

	v1p "github.com/atomreforge/daizy-night-server/internal/api/v1/public/user"
)

func (h *HandlerComplex) HandleNotifGetSingle(ctx *echo.Context) error {
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerNotif))

	req, err := Bind[v1p.RequestGetNotificationSingle](ctx)
	if err != nil {
		return err
	}

	utils.Layer(ctx).Info(fmt.Sprintf(
		"%s; user::%s; nid::%s",
		string(consts.ExprReqNotifGetSingle),
		ctx.Param("username"),
		strconv.FormatUint(uint64(req.Nid), 10),
	))

	// will be 0 if not provided in json
	if req.Nid == 0 {
		//return mid.Respond(ctx, http.StatusBadRequest, "nid is required")
		return errs.BuildErrValidation(
			errs.ValidationKeyNull,
			http.StatusBadRequest,
			string(consts.JsonExprNid),
			string(consts.ExprValueNull),
		)
	}

	// get latest nid from repo
	nidLatest, err := h.serviceNotification.GetLatestNotificationNid()
	if err != nil {
		return err
	}

	// validation
	if req.Nid > nidLatest {
		return errs.BuildErrValidation(
			errs.ValidationKeyBadValue,
			http.StatusBadRequest,
			string(consts.JsonExprNid),
			strconv.FormatUint(uint64(req.Nid), 10),
		)
	}

	n, err := h.serviceNotification.GetNotificationSingelByNid(req.Nid)
	if err != nil {
		return err
	}

	resp := v1p.ResponseGetNotificationSingle{
		Item:      *n,
		NidLatest: nidLatest,
	}

	return mid.RespondObj(ctx, http.StatusOK, resp)
}

//func (h *HandlerComplex) HandleNotifGetBatch(ctx *echo.Context) error

// havent ok
// req admin permission
func (h *HandlerComplex) HandleNotificationPost(ctx *echo.Context, n *model.Notification) error {
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerNotif))
	utils.Layer(ctx).Info("")

	h.serviceNotification.PostNotification(n)

	return mid.Respond(ctx, http.StatusInternalServerError, string(consts.ExprUnsupportedFeature))
}
