package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) FindUsersByFilters(ctx *gin.Context) {
	users, err := h.userManagementService.GetAllUserData(ctx, ctx.Query("firstName"), ctx.Query("surname"),
		ctx.Query("phoneNumber"), ctx.Query("email"), ctx.Query("active"), ctx.Query("verified"), ctx.Query("courseName"),
		ctx.Query("page"), ctx.Query("limit"))

	if err != nil {
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, users)
}
