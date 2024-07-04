package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

func (h Handlers) DeleteAdmin(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "DeleteAdmin", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	if err := h.adminService.EraseAdmin(ctx, ctx.Query("login")); err != nil {
		h.logger.Error("не получилось удалить админа", "DeleteAdmin", err.Message, err.Code)
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

	h.logger.Info("админ упешно удален", "DeleteAdmin", ctx.Query("login"))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("администратор успешно удален"))
}

func (h Handlers) ChangeRole(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "ChangeRole", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	if err := h.adminService.ManageRole(ctx, ctx.Query("login"), ctx.Query("role")); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении роли на %v у админа с логином %v", ctx.Query("role"), ctx.Query("login")), "ChangeRole", err.Message, err.Code)
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

	h.logger.Info(fmt.Sprintf("роль админа с логином %v успешно изменена", ctx.Query("login")), "ChangeRole", ctx.Query("role"))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("роль успешно изменена"))
}

func (h Handlers) FindAdmins(ctx *gin.Context) {
	login := ctx.Query("login")
	role := ctx.Query("role")
	twoStepsAuth := ctx.Query("twoStepsAuth")
	page := ctx.Query("page")
	limit := ctx.Query("limit")
	admins, err := h.adminService.RetreiveAdmins(ctx, login, role, twoStepsAuth, page, limit)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при поиске админов по запросу: login - %v, role - %v, twoStepsAuth - %v, page - %v, limit - %v", login, role, twoStepsAuth, page, limit), "FindAdmins", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("админы успешно найдены", "FindAdmins", fmt.Sprintf("фильтры: login - %v, role - %v, twoStepsAuth - %v, page - %v, limit - %v", login, role, twoStepsAuth, page, limit))

	ctx.JSON(http.StatusOK, admins)
}

func (h Handlers) GetPaymentDashboard(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" && role != "admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "GetPaymentDashboard", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	from := ctx.Query("from")
	due := ctx.Query("due")
	courseName := ctx.Query("courseName")
	paymentMethod := ctx.Query("paymentMethod")

	stats, err := h.adminService.GetPaymentsData(ctx, from, due, courseName, paymentMethod)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при получении статистики по платежам по фильтрам: from -  %v, due - %v, courseName - %v, paymentMethod - %v", from, due, courseName, paymentMethod), "GetPaymentDashboard", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("статистика успешно получена", "GetPaymentDashboard", fmt.Sprintf("фильтры: from -  %v, due - %v, courseName - %v, paymentMethod - %v", from, due, courseName, paymentMethod))

	ctx.JSON(http.StatusOK, stats)
}

func (h Handlers) GetUsersDashboard(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" && role != "admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "GetUsersDashboard", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	stats, err := h.adminService.GetUsersData(ctx, ctx.Query("from"), ctx.Query("due"))
	if err != nil {
		h.logger.Error("ошибка при получении статистики пользователей", "GetUsersDashboard", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("статистика по пользователям успшно получена", "GetUsersDashboard", "")

	ctx.JSON(http.StatusOK, stats)
}
