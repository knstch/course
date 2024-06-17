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
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if courseErr.Code == 11050 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
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
	file, header, err := ctx.Request.FormFile("lesson")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(err, 400))
		return
	}

	_, courseErr := h.contentManagementService.AddLesson(ctx, &file, header)
	if courseErr != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
	}
}
