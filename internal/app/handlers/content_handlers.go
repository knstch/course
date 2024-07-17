package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/knstch/course/internal/app/course_error"
	contentmanagement "github.com/knstch/course/internal/app/services/content_management"
)

// @Summary Найти курсы по фильтрам
// @Produce json
// @Description Используется для получения курсов по фильтрам. При передаче ID все остальные фильтры игнорируются и происходит проверка на наличие доступа к контенту.
// @Success 200 {object} entity.CourseInfoWithPagination
// @Router /v1/content/courses [get]
// @Router /v1/profile/getCourses [get]
// @Router /v1/admin/management/courses [get]
// @Tags Методы взаимодействия с контентом
// @Param id query string false "ID курса"
// @Param name query string false "Название курса"
// @Param description query string false "Описание курса"
// @Param cost query int false "Стоимость курса"
// @Param discount query int false "Размер скидки"
// @Param page query string true "Страница"
// @Param limit query string true "Лимит"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 403 {object} courseerror.CourseError "Нет доступа к расширенному контенту"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) RetreiveCourses(ctx *gin.Context) {
	var statusCode int

	params := contentmanagement.CourseQueryParams{
		ID:          ctx.Query("id"),
		Name:        ctx.Query("name"),
		Description: ctx.Query("description"),
		Cost:        ctx.Query("cost"),
		Discount:    ctx.Query("discount"),
		Page:        ctx.Query("page"),
		Limit:       ctx.Query("limit"),
	}

	coursesInfo, err := h.contentManagementService.GetCourseInfo(ctx, &params)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при получения курсов по фильтрам: id - %v, name - %v, description - %v, cost - %v, discount - %v, page - %v, limit - %v",
			params.ID,
			params.Name,
			params.Description,
			params.Cost,
			params.Discount,
			params.Page,
			params.Limit,
		), "RetreiveCourses", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "RetreiveCourses")
			return
		}
		if err.Code == 13004 {
			statusCode = http.StatusForbidden
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "RetreiveCourses")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "GET", "RetreiveCourses")
		return
	}

	h.logger.Info("курсы успешно получены", "RetreiveCourses",
		fmt.Sprintf("фильтры: id - %s, name - %s, description - %s, cost - %s, discount - %s, page - %s, limit - %s",
			params.ID,
			params.Name,
			params.Description,
			params.Cost,
			params.Discount,
			params.Page,
			params.Limit,
		))

	statusCode = http.StatusOK
	ctx.JSON(statusCode, coursesInfo)
	h.metrics.RecordResponse(statusCode, "GET", "RetreiveCourses")
}

// @Summary Найти модули по фильтрам
// @Description Используется для получения модулей. Если было передано название курса, к которому принадлежит модуль, и пользователь залогинен то происходит проверка на наличие курса в профиле пользователя. Если он не куплен, то возвращается ошибка.
// @Produce json
// @Success 200 {object} entity.ModuleInfoWithPagination
// @Router /v1/content/modules [get]
// @Router /v1/profile/modules [get]
// @Router /v1/admin/management/modules [get]
// @Tags Методы взаимодействия с контентом
// @Param name query string false "Название модуля"
// @Param description query string false "Описание модуля"
// @Param courseName query string false "Название курса"
// @Param page query string true "Страница"
// @Param limit query string true "Лимит"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) RetreiveModules(ctx *gin.Context) {
	var statusCode int

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
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "RetreiveModules")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "GET", "RetreiveModules")
		return
	}

	h.logger.Info("модули успешно получены", "RetreiveModules",
		fmt.Sprintf("фильтры: name - %v, description - %v, courseName - %v, page - %v, limit - %v",
			name, description, courseName, page, limit))

	statusCode = http.StatusOK
	ctx.JSON(statusCode, modules)
	h.metrics.RecordResponse(statusCode, "GET", "RetreiveModules")
}

// @Summary Найти уроки по фильтрам
// @Produce json
// @Success 200 {object} entity.ModuleInfoWithPagination
// @Description Используется для получения уроков. Если было передано название курса, к которому принадлежит модуль, и пользователь залогинен то происходит проверка на наличие курса в профиле пользователя. Если он не куплен, то возвращается ошибка.
// @Router /v1/content/lessons [get]
// @Router /v1/profile/lessons [get]
// @Router /v1/admin/management/lessons [get]
// @Tags Методы взаимодействия с контентом
// @Param name query string false "Название урока"
// @Param description query string false "Описание урока"
// @Param courseName query string false "Название курса"
// @Param moduleName query string false "Название модуля"
// @Param page query string true "Страница"
// @Param limit query string true "Лимит"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) RetreiveLessons(ctx *gin.Context) {
	var statusCode int

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
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "RetreiveLessons")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "GET", "RetreiveLessons")
		return
	}

	h.logger.Info("уроки успешно получены",
		"RetreiveLessons",
		fmt.Sprintf("фильтры: name - %v, description - %v, moduleName - %v, courseName - %v, page - %v, limit - %v",
			name, description, moduleName, courseName, page, limit))

	statusCode = http.StatusOK
	ctx.JSON(statusCode, lessons)
	h.metrics.RecordResponse(statusCode, "GET", "RetreiveLessons")
}
