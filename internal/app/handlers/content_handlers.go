package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h Handlers) RetreiveCourses(ctx *gin.Context) {
	coursesInfo, err := h.contentManagementService.GetCourseInfo(ctx, ctx.Query("id"), ctx.Query("name"), ctx.Query("description"), ctx.Query("cost"), ctx.Query("discount"))
	if err != nil {
		if err.Code == 13003 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 13004 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, coursesInfo)
}
