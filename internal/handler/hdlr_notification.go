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
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"

	v1p "github.com/atomreforge/daizy-night-server/internal/api/v1/public/user"
)

// req admin
func (h *HandlerComplex) HandleNotificationCreate(ctx *echo.Context) error {
	// trace
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerNotif))
	utils.Layer(ctx).Info(string(consts.ExprReqNotifCreate))

	// build req body
	req, err := Bind[v1p.RequestCreateNotification](ctx)
	if err != nil {
		return err
	}

	// build notif; Nid is server-allocated by the repo and PostBy/PostAt are
	// stamped when the notification is posted, so all three stay zero here
	n := model.Notification{
		Md:           req.Md,
		Nid:          0,
		TitlePreview: req.TitlePreview,
		PostBy:       0,
		Levels:       req.Levels,
	}

	// formal repo ops
	err = h.serviceNotification.CreateNotification(n)
	if err != nil {
		return err
	}

	// respond
	return mid.RespondObj(ctx, http.StatusOK, v1p.ResponseCreateNotification{Message: string(consts.JsonExprMessageOk)})
}

// req admin
func (h *HandlerComplex) HandleNotificationDelete(ctx *echo.Context) error {
	// trace
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerNotif))
	utils.Layer(ctx).Info(string(consts.ExprReqNotifDelete))

	// build req body
	req, err := Bind[v1p.RequestDeleteNotification](ctx)
	if err != nil {
		return err
	}

	// validate: nid is a strictly positive integer
	if req.Nid == 0 {
		return errs.BuildErrValidation(
			errs.ValidationKeyNull,
			http.StatusBadRequest,
			string(consts.JsonExprNid),
			string(consts.ExprValueNull),
		)
	}

	// formal repo ops
	err = h.serviceNotification.DeleteNotification(req.Nid)
	if err != nil {
		return err
	}

	// resp
	return mid.RespondObj(ctx, http.StatusOK, v1p.ResponseDeleteNotification{Message: string(consts.JsonExprMessageOk)})
}

// req admin
func (h *HandlerComplex) HandleNotificationUpdate(ctx *echo.Context) error {
	// trace
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerNotif))
	utils.Layer(ctx).Info(string(consts.ExprReqNotifUpdate))

	// build req body
	req, err := Bind[v1p.RequestUpdateNotification](ctx)
	if err != nil {
		return err
	}

	// validate: nid selects the target and is a strictly positive integer
	if req.Nid == 0 {
		return errs.BuildErrValidation(
			errs.ValidationKeyNull,
			http.StatusBadRequest,
			string(consts.JsonExprNid),
			string(consts.ExprValueNull),
		)
	}

	// explicit construction: never forward client-controlled gorm.Model
	// columns (id/created_at/updated_at/deleted_at); PostBy is preserved
	// server-side by the repo's UpdateItem
	n := model.Notification{
		Md:           req.Md,
		Nid:          req.Nid,
		TitlePreview: req.TitlePreview,
		Levels:       req.Levels,
	}

	// repo ops
	err = h.serviceNotification.UpdateNotification(n)
	if err != nil {
		return err
	}

	// resp
	return mid.RespondObj(ctx, http.StatusOK, v1p.ResponseUpdateNotification{Message: string(consts.JsonExprMessageOk)})
}

// req admin
func (h *HandlerComplex) HandleNotificationWithdraw(ctx *echo.Context) error {
	// trace
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerNotif))
	utils.Layer(ctx).Info(string(consts.ExprReqNotifWithdraw))

	// build req body
	req, err := Bind[v1p.RequestWithdrawNotification](ctx)
	if err != nil {
		return err
	}

	// validate: nid is a strictly positive integer
	if req.Nid == 0 {
		return errs.BuildErrValidation(
			errs.ValidationKeyNull,
			http.StatusBadRequest,
			string(consts.JsonExprNid),
			string(consts.ExprValueNull),
		)
	}

	// repo ops
	err = h.serviceNotification.WithdrawNotification(req.Nid)
	if err != nil {
		return err
	}

	// resp
	return mid.RespondObj(ctx, http.StatusOK, v1p.ResponseWithdrawNotification{Message: string(consts.JsonExprMessageOk)})
}

func (h *HandlerComplex) HandleNotificationGetSingle(ctx *echo.Context) error {
	// trace
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerNotif))

	// build req body
	req, err := Bind[v1p.RequestGetNotificationSingle](ctx)
	if err != nil {
		return err
	}

	// trace
	utils.Layer(ctx).Info(fmt.Sprintf(
		"%s; nid::%s",
		string(consts.ExprReqNotifGetSingle),
		strconv.FormatUint(uint64(req.Nid), 10),
	))

	// value will be set to 0 in binding if not provided in json
	if req.Nid == 0 {
		return errs.BuildErrValidation(
			errs.ValidationKeyNull,
			http.StatusBadRequest,
			string(consts.JsonExprNid),
			string(consts.ExprValueNull),
		)
	}

	// call service
	n, err := h.serviceNotification.GetNotificationSingleByNid(req.Nid)
	if err != nil {
		return err
	}

	// build resp body
	resp := v1p.ResponseGetNotificationSingle{
		Payload: *n,
	}

	// http respond
	return mid.RespondObj(ctx, http.StatusOK, resp)
}

func (h *HandlerComplex) HandleNotificationGetBatch(ctx *echo.Context) error {
	// trace
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerNotif))
	utils.Layer(ctx).Info(string(consts.ExprReqNotifGetBatch))

	// build req body
	req, err := Bind[v1p.RequestGetNotificationBatch](ctx)
	if err != nil {
		return err
	}

	// range validation lives in the service layer (single source of truth)

	// call notif service
	batch, err := h.serviceNotification.GetNotificationBatchByNid(req.NidBegin, req.NidEnd)
	if err != nil {
		return err
	}

	// build resp body
	resp := v1p.ResponseGetNotificationBatchUpdate{
		Payload: *batch,
		Message: string(consts.HttpExprOk),
	}

	// http respond
	return mid.RespondObj(ctx, http.StatusOK, resp)

}

func (h *HandlerComplex) HandleNotificationGetBatchUpdate(ctx *echo.Context) error {
	// trace
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerNotif))
	utils.Layer(ctx).Info(string(consts.ExprReqNotifGetUpdateBatch))

	// build req body
	req, err := Bind[v1p.RequestGetNotificationBatchUpdate](ctx)
	if err != nil {
		return err
	}

	// call service
	batch, err := h.serviceNotification.GetNotificationBatchUpdate(req.ClientLatestNid)
	if err != nil {
		return err
	}

	// build resp body
	resp := v1p.ResponseGetNotificationBatch{
		Payload: *batch,
	}

	// http respond
	return mid.RespondObj(ctx, http.StatusOK, resp)
}

// req admin permission;
func (h *HandlerComplex) HandleNotificationPost(ctx *echo.Context) error {
	// trace
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerNotif))
	utils.Layer(ctx).Info(string(consts.ExprReqNotifPost))

	// build req body
	req, err := Bind[v1p.RequestPostNotification](ctx)
	if err != nil {
		return err
	}

	// validate: nid is a strictly positive integer
	if req.Nid == 0 {
		return errs.BuildErrValidation(
			errs.ValidationKeyNull,
			http.StatusBadRequest,
			string(consts.JsonExprNid),
			string(consts.ExprValueNull),
		)
	}

	// acting publisher from the jwt claims
	token, err := echo.ContextGet[*jwt.Token](ctx, string(consts.CtxExprKeyJWT))
	if err != nil {
		return err
	}
	claims, ok := token.Claims.(*model.JwtAccessTokenPayload)
	if !ok {
		return errs.BuildUnauthorized("failed to parse token claims")
	}

	// repo ops: stamp post_by/post_at
	if err := h.serviceNotification.PostNotification(req.Nid, claims.Uid); err != nil {
		return err
	}

	// http respond
	return mid.Respond(ctx, http.StatusOK, string(consts.HttpExprOk))
}
