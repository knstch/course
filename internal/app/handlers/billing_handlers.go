package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

var (
	errNotVerified = errors.New("для покупки курса нужно верифицировать почту")
)

func (h Handlers) BuyCourse(ctx *gin.Context) {
	verifiedStatus := ctx.GetBool("verified")
	if !verifiedStatus {
		h.logger.Error(fmt.Sprintf("пользователь не верифицирован, ID: %d", ctx.Value("userId").(uint)), "BuyCourse", errNotVerified.Error(), 11008)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNotVerified, 11008))
	}

	buyDetails := entity.CreateNewBuyDetails()
	if err := ctx.ShouldBindJSON(&buyDetails); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "BuyCourse", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	linkToPay, err := h.sberBillingService.PlaceOrder(ctx, buyDetails)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при размещении заказа пользователя с ID: %d при покупке курса с ID: %d", ctx.Value("userId"), buyDetails.CourseId), "BuyCourse", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 11002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 15004 {
			ctx.AbortWithStatusJSON(http.StatusConflict, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("заказ пользователя с ID: %d был успешно размещен", ctx.Value("userId")), "BuyCourse", fmt.Sprint(buyDetails.CourseId))

	ctx.Redirect(http.StatusTemporaryRedirect, *linkToPay)
}

func (h Handlers) CompletePurchase(ctx *gin.Context) {
	courseName, err := h.sberBillingService.ConfirmPayment(ctx, ctx.Param("userData"))
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при завершении покупки заказа %v", ctx.Param("userData")), "CompletePurchase", err.Message, err.Code)
		if err.Code == 11004 || err.Code == 15001 || err.Code == 15002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 15003 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("курс успешно приобретен", "CompletePurchase", ctx.Param("userData"))

	ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%v/%v", h.address, courseName))
}

func (h Handlers) DeclineOrder(ctx *gin.Context) {
	if err := h.sberBillingService.FailPayment(ctx, ctx.Param("userData")); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при отмененной оплатае с заказом: %v", ctx.Param("userData")), "DeclineOrder", err.Message, err.Code)
		if err.Code == 11004 || err.Code == 15001 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("заказ успешно отменен", "DeclineOrder", ctx.Param("userData"))

	ctx.JSON(http.StatusOK, entity.CreateNewFailedPayment)
}

func (h Handlers) ManageBillingHost(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "ManageBillingHost", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	host := entity.CreateBillingHost()
	if err := ctx.ShouldBindJSON(&host); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "ManageBillingHost", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.sberBillingService.ChangeApiHost(ctx, host.Url); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении BillingApiHost на %v", host.Url), "ManageBillingHost", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("BillingApiHost успешно изменен админом с ID: %d", ctx.Value("adminId")), "ManageBillingHost", host.Url)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("хост успешно изменен"))
}

func (h Handlers) ManageAccessToken(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "ManageAccessToken", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	token := entity.CreateAccessToken()
	if err := ctx.ShouldBindJSON(&token); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "ManageAccessToken", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.sberBillingService.ChangeAccessToken(ctx, token.Token); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении токена доступа для billingHost на токен: %v", token.Token), "ManageAccessToken", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("токен billingHost успешно изменен админом с ID: %d", ctx.Value("adminId")), "ManageAccessToken", token.Token)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("токен успешно изменен"))
}
