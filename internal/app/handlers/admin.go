package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

func (h Handlers) DeleteAdmin(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	if err := h.adminService.EraseAdmin(ctx, ctx.Query("login")); err != nil {
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 16002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("администратор успешно удален", true))
}

func (h Handlers) ChangeRole(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	if err := h.adminService.ManageRole(ctx, ctx.Query("login"), ctx.Query("role")); err != nil {
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 16002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("роль успешно изменена", true))
}

func (h Handlers) FindAdmins(ctx *gin.Context) {
	admins, err := h.adminService.RetreiveAdmins(ctx, ctx.Query("login"), ctx.Query("role"), ctx.Query("twoStepsAuth"), ctx.Query("page"), ctx.Query("limit"))
	if err != nil {
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, admins)
}

func (h Handlers) GetPaymentDashboard(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" && role != "admin" {
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	stats, err := h.adminService.GetPaymentData(ctx, ctx.Query("from"), ctx.Query("due"), ctx.Query("courseName"), ctx.Query("paymentMethod"))
	if err != nil {
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
