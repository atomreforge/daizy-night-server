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

// calendarClaims resolves the authenticated access-token claims; every
// user-scoped handler needs them for the ownership check.
func calendarClaims(ctx *echo.Context) (*model.JwtAccessTokenPayload, error) {
	token, err := echo.ContextGet[*jwt.Token](ctx, string(consts.CtxExprKeyJWT))
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*model.JwtAccessTokenPayload)
	if !ok {
		utils.Layer(ctx).Error("failed to assert claims to JwtAccessTokenPayload")
		return nil, echo.ErrUnauthorized
	}
	return claims, nil
}

func (h *HandlerComplex) HandleCalendarGet(ctx *echo.Context) error {
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerCalendar))

	d := GetDomain(ctx)
	switch d {
	// no Attr-Roaming in public domain
	case consts.DomainPublic:
		utils.Layer(ctx).Info(
			fmt.Sprintf(
				"%s; user::%s; domain::%s",
				consts.ExprReqCalendarGet,
				ctx.Param("username"),
				d.Say()),
		)
		cal, err := h.ServiceUser.GetCalendarByUsername(ctx.Param("username"))
		utils.AppendCallChain(ctx, string(consts.ModExprServiceUser))
		if err != nil {
			return err
		}

		out := v1public.ResponseAnyCalendarGet{
			UserID:     cal.UserID,
			CalendarID: cal.CalendarID,
			Records:    []v1public.ResponseAnyCalendarItem{},
		}
		for _, it := range cal.Records {
			out.Records = append(out.Records, v1public.ResponseAnyCalendarItem{
				CalendarID: it.CalendarID,
				Weekday:    it.Weekday,
				StartMin:   it.StartMin,
				EndMin:     it.EndMin,
				Title:      it.Title,
			})
		}
		return mid.RespondObj(ctx, http.StatusOK, out)

	case consts.DomainPrivate:
		utils.Layer(ctx).Info(
			fmt.Sprintf(
				"%s; user::%s; domain::%s",
				consts.ExprReqCalendarGet,
				ctx.Param("username"),
				d.Say()),
		)

		claims, err := calendarClaims(ctx)
		if err != nil {
			return err
		}
		if err := requireSelf(ctx, claims); err != nil {
			return err
		}

		cal, err := h.ServiceUser.GetCalendarByUid(claims.Uid)
		utils.AppendCallChain(ctx, string(consts.ModExprServiceUser))
		if err != nil {
			return err
		}

		out := v1.ResponseUserCalendarGet{
			Roaming:    cal.Roaming,
			UserID:     cal.UserID,
			CalendarID: cal.CalendarID,
			Records:    []v1.ResponseUserCalendarItem{},
		}
		for _, it := range cal.Records {
			out.Records = append(out.Records, v1.ResponseUserCalendarItem{
				Roaming:    it.Roaming,
				CalendarID: it.CalendarID,
				Weekday:    it.Weekday,
				StartMin:   it.StartMin,
				EndMin:     it.EndMin,
				Title:      it.Title,
			})
		}
		return mid.RespondObj(ctx, http.StatusOK, out)
	}
	return mid.Respond(ctx, http.StatusBadRequest, string(consts.ExprUnreachableCase))
}

func (h *HandlerComplex) HandleCalendarPut(ctx *echo.Context) error {
	// record flow chain (monotonically accumulating)
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerCalendar))
	if GetDomain(ctx) != consts.DomainPrivate {
		return echo.ErrForbidden
	}

	b, err := Bind[model.CalendarPutBody](ctx)
	if err != nil {
		return err
	}
	utils.Layer(ctx).Info(fmt.Sprintf("%s; records::%d", consts.ExprReqCalendarPut, len(b.Records)))

	// failure during param validation
	if err := ValidateCalendarPutParams(&b); err != nil {
		return err
	}

	claims, err := calendarClaims(ctx)
	if err != nil {
		return err
	}
	if err := requireSelf(ctx, claims); err != nil {
		return err
	}

	// the owner comes from the token, never from the request body
	cal := model.CalendarTable{UserID: claims.Uid}
	for _, it := range b.Records {
		cal.Records = append(cal.Records, model.CalendarItem{
			Weekday:  it.Weekday,
			StartMin: it.StartMin,
			EndMin:   it.EndMin,
			Title:    it.Title,
		})
	}

	// execute service
	utils.AppendCallChain(ctx, string(consts.ModExprServiceUser))
	if err := h.ServiceUser.UpdateCalendar(&cal); err != nil {
		return err
	}
	return mid.RespondObj(ctx, http.StatusOK, v1.ResponseUserCalendarPut{Message: string(consts.HttpExprOk)})
}

func (h *HandlerComplex) HandleCalendarDelete(ctx *echo.Context) error {
	// record flow chain (monotonically accumulating)
	utils.AppendCallChain(ctx, string(consts.ModExprHandlerCalendar))
	utils.Layer(ctx).Info(fmt.Sprintf("%s", consts.ExprReqCalendarDelete))
	if GetDomain(ctx) != consts.DomainPrivate {
		return echo.ErrForbidden
	}

	claims, err := calendarClaims(ctx)
	if err != nil {
		return err
	}
	if err := requireSelf(ctx, claims); err != nil {
		return err
	}

	// execute service
	utils.AppendCallChain(ctx, string(consts.ModExprServiceUser))
	if err := h.ServiceUser.RemoveCalendarByUid(claims.Uid); err != nil {
		return err
	}
	return mid.RespondObj(ctx, http.StatusOK, v1.ResponseUserCalendarDelete{Message: string(consts.HttpExprOk)})
}
