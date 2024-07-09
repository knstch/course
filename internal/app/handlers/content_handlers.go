package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Найти курсы по фильтрам
// @Produce json
// @Success 200 {object} entity.CourseInfoWithPagination
// @Router /v1/content/courses [get]
// @Router /v1/profile/getCourses [get]
// @Router /v1/admin/management/courses [get]
// @Tags Методы взаимодействия с контентом
// @Param id query string false "ID"
// @Param name query string false "название курса"
// @Param description query string false "описание курса"
// @Param cost query int false "стоимость курса"
// @Param discount query int false "размер скидки"
// @Param page query string true "страница"
// @Param limit query string true "лимит"
// @Failure 400 {object} courseError.CourseError "Провалена валидация"
// @Failure 403 {object} courseError.CourseError "Нет доступа к расширенному контенту"
// @Failure 500 {object} courseError.CourseError "Возникла внутренняя ошибка"
func (h Handlers) RetreiveCourses(ctx *gin.Context) {
	id := ctx.Query("id")
	name := ctx.Query("name")
	description := ctx.Query("description")
	cost := ctx.Query("cost")
	discount := ctx.Query("discount")
	page := ctx.Query("page")
	limit := ctx.Query("limit")
	coursesInfo, err := h.contentManagementService.GetCourseInfo(
		ctx,
		id,
		name,
		description,
		cost,
		discount,
		page,
		limit,
	)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при получения курсов по фильтрам: id - %v, name - %v, description - %v, cost - %v, discount - %v, page - %v, limit - %v",
			id,
			name,
			description,
			cost,
			discount,
			page,
			limit), "RetreiveCourses", err.Message, err.Code)
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

	h.logger.Info("курсы успешно получены", "RetreiveCourses",
		fmt.Sprintf("фильтры: id - %v, name - %v, description - %v, cost - %v, discount - %v, page - %v, limit - %v", id,
			name,
			description,
			cost,
			discount,
			page,
			limit))

	ctx.JSON(http.StatusOK, coursesInfo)
}

// @Summary Найти модули по фильтрам
// @Produce json
// @Success 200 {object} entity.ModuleInfoWithPagination
// @Router /v1/content/modules [get]
// @Router /v1/profile/modules [get]
// @Router /v1/admin/management/modules [get]
// @Tags Методы взаимодействия с контентом
// @Param name query string false "название модуля"
// @Param description query string false "описание модуля"
// @Param courseName query string false "название курса"
// @Param page query string true "страница"
// @Param limit query string true "лимит"
// @Failure 400 {object} courseError.CourseError "Провалена валидация"
// @Failure 500 {object} courseError.CourseError "Возникла внутренняя ошибка"
func (h Handlers) RetreiveModules(ctx *gin.Context) {
	name := ctx.Query("name")
	description := ctx.Query("description")
	courseName := ctx.Query("courseName")
	page := ctx.Query("page")
	limit := ctx.Query("limit")
	modules, err := h.contentManagementService.GetModulesInfo(
		ctx,
		name,
		description,
		courseName,
		page,
		limit)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при получении модулей по фильтрам: name - %v, description - %v, courseName - %v, page - %v, limit - %v", name, description, courseName, page, limit),
			"RetreiveModules", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("модули успешно получены", "RetreiveModules",
		fmt.Sprintf("фильтры: name - %v, description - %v, courseName - %v, page - %v, limit - %v",
			name, description, courseName, page, limit))

	ctx.JSON(http.StatusOK, modules)
}

// @Summary Найти уроки по фильтрам
// @Produce json
// @Success 200 {object} entity.ModuleInfoWithPagination
// @Router /v1/content/lessons [get]
// @Router /v1/profile/lessons [get]
// @Router /v1/admin/management/lessons [get]
// @Tags Методы взаимодействия с контентом
// @Param name query string false "название урока"
// @Param description query string false "описание урлка"
// @Param courseName query string false "название курса"
// @Param moduleName query string false "название модуля"
// @Param page query string true "страница"
// @Param limit query string true "лимит"
// @Failure 400 {object} courseError.CourseError "Провалена валидация"
// @Failure 500 {object} courseError.CourseError "Возникла внутренняя ошибка"
func (h Handlers) RetreiveLessons(ctx *gin.Context) {
	name := ctx.Query("name")
	description := ctx.Query("description")
	moduleName := ctx.Query("moduleName")
	courseName := ctx.Query("courseName")
	page := ctx.Query("page")
	limit := ctx.Query("limit")
	lessons, err := h.contentManagementService.GetLessonsInfo(ctx,
		ctx.Query("name"),
		ctx.Query("description"),
		ctx.Query("moduleName"),
		ctx.Query("courseName"),
		ctx.Query("page"),
		ctx.Query("limit"))
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка получения уроков по фильтрам: name - %v, description - %v, moduleName - %v, courseName - %v, page - %v, limit - %v",
			name, description, moduleName, courseName, page, limit), "RetreiveLessons", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("уроки успешно получены",
		"RetreiveLessons",
		fmt.Sprintf("фильтры: name - %v, description - %v, moduleName - %v, courseName - %v, page - %v, limit - %v",
			name, description, moduleName, courseName, page, limit))

	ctx.JSON(http.StatusOK, lessons)
}
