package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

func (h Handlers) FindUsersByFilters(ctx *gin.Context) {
	users, err := h.userManagementService.RetreiveUsersByFilters(ctx, ctx.Query("firstName"), ctx.Query("surname"),
		ctx.Query("phoneNumber"), ctx.Query("email"), ctx.Query("active"), ctx.Query("verified"), ctx.Query("courseName"),
		ctx.Query("banned"), ctx.Query("page"), ctx.Query("limit"))

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

func (h Handlers) BanUser(ctx *gin.Context) {
	Id := entity.NewId(nil)
	if err := ctx.ShouldBindJSON(&Id); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.userManagementService.DeactivateUser(ctx, Id.Id); err != nil {
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пользователь успешно заблокирован", true))
}

func (h Handlers) GetUserById(ctx *gin.Context) {
	user, err := h.userManagementService.RetreiveUserById(ctx, ctx.Query("id"))
	if err != nil {
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 11101 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}
