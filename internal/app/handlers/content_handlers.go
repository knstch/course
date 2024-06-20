package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h Handlers) RetreiveCourses(ctx *gin.Context) {
	coursesInfo, err := h.contentManagementService.GetCourseInfo(
		ctx, ctx.Query("id"),
		ctx.Query("name"),
		ctx.Query("description"),
		ctx.Query("cost"),
		ctx.Query("discount"),
		ctx.Query("page"),
		ctx.Query("limit"),
	)
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

func (h Handlers) RetreiveModules(ctx *gin.Context) {
	modules, err := h.contentManagementService.GetModulesInfo(
		ctx,
		ctx.Query("name"),
		ctx.Query("description"),
		ctx.Query("courseName"),
		ctx.Query("page"),
		ctx.Query("limit"))
	if err != nil {
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, modules)
}

func (h *Handlers) RetreiveLessons(ctx *gin.Context) {
	lessons, err := h.contentManagementService.GetLessonsInfo(ctx,
		ctx.Query("name"),
		ctx.Query("description"),
		ctx.Query("moduleName"),
		ctx.Query("courseName"),
		ctx.Query("page"),
		ctx.Query("limit"))
	if err != nil {
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, lessons)
}
