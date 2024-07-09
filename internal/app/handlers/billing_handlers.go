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

// @Summary Купить курс
// @Accept json
// @Success 307 "Temporary Redirect"
// @Router /v1/billing/buyCourse [post]
// @Tags Методы биллинга
// @Param orderDetails body entity.BuyDetails true "ID курса и способ платежа"
// @Failure 400 {object} courseError.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 409 {object} courseError.CourseError "Курс уже куплен"
// @Failure 500 {object} courseError.CourseError "Возникла внутренняя ошибка"
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

// @Summary Оплата курса подтверждена
// @Success 307 "Temporary Redirect"
// @Router /v1/billing/successPayment/{userData} [get]
// @Tags Методы биллинга
// @Param userData path string true "захешированные данные пользователя"
// @Failure 400 {object} courseError.CourseError "Инвойс ID не совпадает с хэшем из path"
// @Failure 404 {object} courseError.CourseError "Заказ не найден"
// @Failure 500 {object} courseError.CourseError "Возникла внутренняя ошибка"
func (h Handlers) CompletePurchase(ctx *gin.Context) {
	userData := ctx.Param("userData")
	courseName, err := h.sberBillingService.ConfirmPayment(ctx, userData)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при завершении покупки заказа %v", userData), "CompletePurchase", err.Message, err.Code)
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

	h.logger.Info("курс успешно приобретен", "CompletePurchase", userData)

	ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%v/%v", h.address, courseName))
}

// @Summary Оплата курса провалена
// @Success 307 "Temporary Redirect"
// @Router /v1/billing/failPayment/{userData} [get]
// @Tags Методы биллинга
// @Param userData path string true "захешированные данные пользователя"
// @Failure 404 {object} courseError.CourseError "Заказ не найден"
// @Failure 500 {object} courseError.CourseError "Возникла внутренняя ошибка"
func (h Handlers) DeclineOrder(ctx *gin.Context) {
	userData := ctx.Param("userData")
	if err := h.sberBillingService.FailPayment(ctx, userData); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при отмененной оплатае с заказом: %v", userData), "DeclineOrder", err.Message, err.Code)
		if err.Code == 11004 || err.Code == 15001 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("заказ успешно отменен", "DeclineOrder", userData)

	ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%v/orders", h.address))
}

// @Summary Изменить billing host
// @Produce json
// @Accept json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/manageBillingHost [patch]
// @Tags Методы биллинга
// @Param billingHost body entity.BillingHost true "Новый URL биллинг хоста"
// @Failure 400 {object} courseError.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 500 {object} courseError.CourseError "Возникла внутренняя ошибка"
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

// @Summary Изменить токен биллинга
// @Accept json
// @Produce json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/manageBillingToken [patch]
// @Tags Методы биллинга
// @Param token query string true "токен"
// @Failure 400 {object} courseError.CourseError "Провалена валидация"
// @Failure 500 {object} courseError.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ManageAccessToken(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "ManageAccessToken", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	token := ctx.Query("token")

	if err := h.sberBillingService.ChangeAccessToken(ctx, token); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении токена доступа для billingHost на токен: %v", token), "ManageAccessToken", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("токен billingHost успешно изменен админом с ID: %d", ctx.Value("adminId")), "ManageAccessToken", token)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("токен успешно изменен"))
}
