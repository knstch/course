package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

// @Summary Удалить админа
// @Produce json
// @Success 200 {object} entity.SuccessResponse
// @Description Используется для удаления админа. Метод доступен только супер админу.
// @Router /v1/admin/management/removeAdmin [delete]
// @Tags Методы для администрирования
// @Param login query string true "логин"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 403 {object} courseerror.CourseError "Нет прав"
// @Failure 404 {object} courseerror.CourseError "Админ не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) DeleteAdmin(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "DeleteAdmin", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseerror.CreateError(errNoRights, 16004))
		return
	}
	login := ctx.Query("login")
	if err := h.adminService.EraseAdmin(ctx, login); err != nil {
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

	h.logger.Info("админ упешно удален", "DeleteAdmin", login)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("администратор успешно удален"))
}

// @Summary Изменить роль администратора
// @Produce json
// @Description Используется для изменения роли админа. Метод доступен только супер админу.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/admin/management/changeRole [patch]
// @Tags Методы для администрирования
// @Param login query string true "логин"
// @Param role query string true "роль"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 403 {object} courseerror.CourseError "Не хватает прав"
// @Failure 404 {object} courseerror.CourseError "Администратор не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ChangeRole(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "ChangeRole", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseerror.CreateError(errNoRights, 16004))
		return
	}
	login := ctx.Query("login")
	adminRole := ctx.Query("role")
	if err := h.adminService.ManageRole(ctx, login, adminRole); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении роли на %v у админа с логином %v", adminRole, login), "ChangeRole", err.Message, err.Code)
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

	h.logger.Info(fmt.Sprintf("роль админа с логином %v успешно изменена", login), "ChangeRole", adminRole)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("роль успешно изменена"))
}

// @Summary Найти администраторов по фильтрам
// @Produce json
// @Description Используется для получения списка администраторов. Метод доступен только супер админу.
// @Success 200 {object} entity.AdminsInfoWithPagination
// @Router /v1/admin/management/getAdmins [get]
// @Tags Методы для администрирования
// @Param login query string false "Логин"
// @Param role query string false "Роль"
// @Param twoStepsAuth query string false "Подключенная двойная авторизация"
// @Param page query string true "Страница"
// @Param limit query string true "Лимит"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 403 {object} courseerror.CourseError "Нет прав"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) FindAdmins(ctx *gin.Context) {
	currentRole := ctx.Value("role").(string)
	if currentRole != "super_admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "ChangeRole", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseerror.CreateError(errNoRights, 16004))
		return
	}

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

// @Summary Получить данные с платежами по дням
// @Produce json
// @Description Используется для получения данных о платежах по дням. Метод доступен только супер админу.
// @Success 200 {object} []entity.PaymentStats
// @Router /v1/admin/management/paymentStats [get]
// @Tags Методы для администрирования
// @Param from query string true "Приод от"
// @Param due query string false "Период до"
// @Param courseName query string false "Название курса"
// @Param paymentMethod query string false "Способ платежа"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 403 {object} courseerror.CourseError "Нет прав"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) GetPaymentDashboard(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" && role != "admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "GetPaymentDashboard", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseerror.CreateError(errNoRights, 16004))
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

// @Summary Получить данные с юзерами по дням
// @Produce json
// @Description Используется для получения данных по новым пользователям по дням. Метод доступен только супер админу.
// @Success 200 {object} []entity.UsersStats
// @Router /v1/admin/management/usersStats [get]
// @Tags Методы для администрирования
// @Param from query string true "Приод от"
// @Param due query string false "Период до"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 403 {object} courseerror.CourseError "Нет прав"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) GetUsersDashboard(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" && role != "admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "GetUsersDashboard", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseerror.CreateError(errNoRights, 16004))
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
