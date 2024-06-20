package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

var (
	errBadFormData = errors.New("одно или несколько полей не заполнены")
)

func (h *Handlers) CreateNewCourse(ctx *gin.Context) {
	name := ctx.PostForm("name")
	description := ctx.PostForm("description")
	cost := ctx.PostForm("cost")
	discount := ctx.PostForm("discount")
	file, header, err := ctx.Request.FormFile("preview")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBadFormData, 400))
		return
	}

	id, courseErr := h.contentManagementService.AddCourse(ctx, name, description, cost, discount, header, &file)
	if courseErr != nil {
		if courseErr.Code == 400 || courseErr.Code == 11105 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, courseErr)
			return
		}
		if courseErr.Code == 11050 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, courseErr)
			return
		}
		if courseErr.Code == 11041 {
			ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, courseErr)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, courseErr)
		return
	}

	ctx.JSON(http.StatusOK, entity.NewId().AddId(id))
}

func (h *Handlers) CreateNewModule(ctx *gin.Context) {
	module := entity.NewModule()
	if err := ctx.ShouldBindJSON(&module); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	id, err := h.contentManagementService.AddModule(ctx, module)
	if err != nil {
		if err.Code == 400 || err.Code == 13001 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, entity.NewId().AddId(id))
}

func (h *Handlers) UploadNewLesson(ctx *gin.Context) {
	lesson, err := ctx.FormFile("lesson")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(err, 400))
		return
	}

	preview, previewHeader, err := ctx.Request.FormFile("preview")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBadFormData, 400))
		return
	}

	name := ctx.PostForm("name")
	moduleName := ctx.PostForm("moduleName")
	description := ctx.PostForm("description")
	position := ctx.PostForm("position")
	courseName := ctx.PostForm("courseName")

	lessonId, courseErr := h.contentManagementService.AddLesson(ctx, lesson, name, moduleName, description, position, courseName, previewHeader, &preview)
	if courseErr != nil {
		if courseErr.Code == 400 || courseErr.Code == 13001 || courseErr.Code == 13002 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, courseErr)
			return
		}
		if courseErr.Code == 11051 || courseErr.Code == 14002 || courseErr.Code == 11041 {
			ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, courseErr)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, courseErr)
		return
	}

	ctx.JSON(http.StatusOK, entity.NewId().AddId(lessonId))
}