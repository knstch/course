package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

var (
	errNotVerified = errors.New("для покупки курса нужно верифицировать почту")
)

// @Summary Купить курс
// @Accept json
// @Description Используется для покупки курса. Формирует инвойс и отправляет его в биллинг. Метод редиректит на страницу оплаты.
// @Success 307 "Temporary Redirect"
// @Router /v1/billing/buyCourse [post]
// @Tags Методы биллинга
// @Param orderDetails body entity.BuyDetails true "ID курса и способ платежа"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 409 {object} courseerror.CourseError "Курс уже куплен"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) BuyCourse(ctx *gin.Context) {
	var statusCode int

	userId := ctx.Value("UserId").(uint)
	verifiedStatus := ctx.GetBool("verified")
	if !verifiedStatus {
		statusCode = http.StatusForbidden
		h.logger.Error(fmt.Sprintf("пользователь не верифицирован, ID: %d", userId), "BuyCourse", errNotVerified.Error(), 11008)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errNotVerified, 11008))
		h.metrics.RecordResponse(statusCode, "POST", "BuyCourse")
		return
	}

	buyDetails := entity.CreateNewBuyDetails()
	if err := ctx.ShouldBindJSON(&buyDetails); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать тело запроса", "BuyCourse", err.Error(), 10101)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.metrics.RecordResponse(statusCode, "POST", "BuyCourse")
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Minute*15)
	defer cancel()

	linkToPay, err := h.sberBillingService.PlaceOrder(timeoutCtx, buyDetails)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при размещении заказа пользователя с ID: %d при покупке курса с ID: %d", userId, buyDetails.CourseId), "BuyCourse", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "BuyCourse")
			return
		}
		if err.Code == 15004 {
			statusCode = http.StatusConflict
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "BuyCourse")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "POST", "BuyCourse")
		return
	}

	h.logger.Info(fmt.Sprintf("заказ пользователя с ID: %d был успешно размещен", userId), "BuyCourse", fmt.Sprint(buyDetails.CourseId))

	statusCode = http.StatusTemporaryRedirect
	ctx.Redirect(statusCode, *linkToPay)
	h.metrics.RecordResponse(statusCode, "POST", "BuyCourse")
}

// @Summary Оплата курса подтверждена
// @Description Используется для подтверждения оплаты платежным шлюзом. Если оплата прошла успешно, редиректит на этот хендлер и затем происходит редирект на страницу с курсом.
// @Success 307 "Temporary Redirect"
// @Router /v1/billing/successPayment/{userData} [get]
// @Tags Методы биллинга
// @Param userData path string true "Захешированные данные пользователя"
// @Failure 400 {object} courseerror.CourseError "Инвойс ID не совпадает с хэшем из path"
// @Failure 404 {object} courseerror.CourseError "Заказ не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) CompletePurchase(ctx *gin.Context) {
	var statusCode int

	userData := ctx.Param("userData")
	courseName, err := h.sberBillingService.ConfirmPayment(ctx, userData)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при завершении покупки заказа %v", userData), "CompletePurchase", err.Message, err.Code)
		if err.Code == 11004 || err.Code == 15001 || err.Code == 15002 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "CompletePurchase")
			return
		}
		if err.Code == 15003 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "CompletePurchase")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "GET", "CompletePurchase")
		return
	}

	h.logger.Info("курс успешно приобретен", "CompletePurchase", userData)

	statusCode = http.StatusTemporaryRedirect
	ctx.Redirect(statusCode, fmt.Sprintf("%v/%v", h.address, courseName))
	h.metrics.RecordResponse(statusCode, "GET", "CompletePurchase")
}

// @Summary Оплата курса провалена
// @Description Используется для отмены заказа, если платеж был провален. Метод редиректит на страницу с заказами пользователя.
// @Success 307 "Temporary Redirect"
// @Router /v1/billing/failPayment/{userData} [get]
// @Tags Методы биллинга
// @Param userData path string true "Захешированные данные пользователя"
// @Failure 404 {object} courseerror.CourseError "Заказ не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) DeclineOrder(ctx *gin.Context) {
	var statusCode int

	userData := ctx.Param("userData")
	if err := h.sberBillingService.FailPayment(ctx, userData); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при отмененной оплатае с заказом: %v", userData), "DeclineOrder", err.Message, err.Code)
		if err.Code == 11004 || err.Code == 15001 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "DeclineOrder")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "GET", "DeclineOrder")
		return
	}

	h.logger.Info("заказ успешно отменен", "DeclineOrder", userData)

	statusCode = http.StatusTemporaryRedirect
	ctx.Redirect(statusCode, fmt.Sprintf("%v/orders", h.address))
	h.metrics.RecordResponse(statusCode, "GET", "DeclineOrder")
}

// @Summary Изменить billing host
// @Produce json
// @Accept json
// @Description Используется для изменения хоста платежного шлюза.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/manageBillingHost [patch]
// @Tags Методы биллинга
// @Param billingHost body entity.BillingHost true "Новый URL биллинг хоста"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ManageBillingHost(ctx *gin.Context) {
	var statusCode int

	role := ctx.Value("Role").(string)
	if role != "super_admin" {
		statusCode = http.StatusForbidden
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("AdminId")), "ManageBillingHost", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errNoRights, 16004))
		h.metrics.RecordResponse(statusCode, "PATCH", "ManageBillingHost")
		return
	}

	host := entity.CreateBillingHost()
	if err := ctx.ShouldBindJSON(&host); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать тело запроса", "ManageBillingHost", err.Error(), 10101)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.metrics.RecordResponse(statusCode, "PATCH", "ManageBillingHost")
		return
	}

	if err := h.sberBillingService.ChangeApiHost(ctx, host.Url); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении BillingApiHost на %v", host.Url), "ManageBillingHost", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "ManageBillingHost")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "PATCH", "ManageBillingHost")
		return
	}

	h.logger.Info(fmt.Sprintf("BillingApiHost успешно изменен админом с ID: %d", ctx.Value("AdminId")), "ManageBillingHost", host.Url)

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("хост успешно изменен"))
	h.metrics.RecordResponse(statusCode, "PATCH", "ManageBillingHost")
}

// @Summary Изменить токен биллинга
// @Accept json
// @Produce json
// @Description Используется для изменения токена доступа к платежному шлюзу.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/manageBillingToken [patch]
// @Tags Методы биллинга
// @Param token query string true "Токен"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ManageAccessToken(ctx *gin.Context) {
	var statusCode int

	role := ctx.Value("Role").(string)
	if role != "super_admin" {
		statusCode = http.StatusForbidden
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("AdminId")), "ManageAccessToken", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errNoRights, 16004))
		h.metrics.RecordResponse(statusCode, "PATCH", "ManageAccessToken")
		return
	}

	token := ctx.Query("token")

	if err := h.sberBillingService.ChangeAccessToken(ctx, token); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении токена доступа для billingHost на токен: %v", token), "ManageAccessToken", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "ManageAccessToken")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "PATCH", "ManageAccessToken")
		return
	}

	h.logger.Info(fmt.Sprintf("токен billingHost успешно изменен админом с ID: %d", ctx.Value("AdminId")), "ManageAccessToken", token)

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("токен успешно изменен"))
	h.metrics.RecordResponse(statusCode, "PATCH", "ManageAccessToken")
}
