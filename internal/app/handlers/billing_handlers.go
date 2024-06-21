package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

func (h Handlers) BuyCourse(ctx *gin.Context) {
	buyDetails := entity.CreateNewBuyDetails()
	if err := ctx.ShouldBindJSON(&buyDetails); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	linkToPay, err := h.sberBillingService.PlaceOrder(ctx, buyDetails)
	if err != nil {
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

	ctx.Redirect(http.StatusTemporaryRedirect, *linkToPay)
}

func (h Handlers) CompletePurchase(ctx *gin.Context) {
	courseName, err := h.sberBillingService.ConfirmPayment(ctx, ctx.Param("userData"))
	if err != nil {
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

	ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%v/%v", h.address, courseName))
}

func (h Handlers) DeclineOrder(ctx *gin.Context) {
	if err := h.sberBillingService.FailPayment(ctx, ctx.Param("userData")); err != nil {
		if err.Code == 11004 || err.Code == 15001 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, entity.CreateNewFailedPayment)
}
