package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

var (
	errBadFormData = errors.New("одно или несколько полей не заполнены")
)

// @Summary Создать курс
// @Accept mpfd
// @Produce json
// @Description Используется для создания нового курса. Требуется токен администратора.
// @Success 200 {object} entity.Id
// @Router /v1/billing/management/createCourse [post]
// @Tags Методы взаимодействия с контентом
// @Param name formData string true "Название курса"
// @Param description formData string true "Описание курса"
// @Param cost formData int true "Стоимость курса"
// @Param discount formData int false "Скидка"
// @Param preview formData file true "Превью"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или неверно передано превью фото"
// @Failure 403 {object} courseerror.CourseError "Ошибка авторизации в CDN"
// @Failure 409 {object} courseerror.CourseError "Курс с таким названием уже существует"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
// @Failure 503 {object} courseerror.CourseError "CDN недоступен"
func (h Handlers) CreateNewCourse(ctx *gin.Context) {
	var statusCode int

	name := ctx.PostForm("name")
	description := ctx.PostForm("description")
	cost := ctx.PostForm("cost")
	discount := ctx.PostForm("discount")
	file, header, err := ctx.Request.FormFile("preview")
	if err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать фото", "CreateNewCourse", err.Error(), 400)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBadFormData, 400))
		h.metrics.RecordResponse(statusCode, "POST", "CreateNewCourse")
		return
	}

	id, courseErr := h.contentManagementService.AddCourse(ctx, name, description, cost, discount, header, &file)
	if courseErr != nil {
		h.logger.Error("не получилось добавить курс", "CreateNewCourse", courseErr.Message, courseErr.Code)
		if courseErr.Code == 400 || courseErr.Code == 11105 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, courseErr)
			h.metrics.RecordResponse(statusCode, "POST", "CreateNewCourse")
			return
		}
		if courseErr.Code == 11050 {
			statusCode = http.StatusForbidden
			ctx.AbortWithStatusJSON(statusCode, courseErr)
			h.metrics.RecordResponse(statusCode, "POST", "CreateNewCourse")
			return
		}
		if courseErr.Code == 11041 {
			statusCode = http.StatusServiceUnavailable
			ctx.AbortWithStatusJSON(statusCode, courseErr)
			h.metrics.RecordResponse(statusCode, "POST", "CreateNewCourse")
			return
		}
		if courseErr.Code == 13001 {
			statusCode = http.StatusConflict
			ctx.AbortWithStatusJSON(statusCode, courseErr)
			h.metrics.RecordResponse(statusCode, "POST", "CreateNewCourse")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, courseErr)
		h.metrics.RecordResponse(statusCode, "POST", "CreateNewCourse")
		return
	}

	h.logger.Info(fmt.Sprintf("курс был успешно размещен админом с ID: %d", ctx.Value("AdminId").(uint)),
		"CreateNewCourse", fmt.Sprintf("courseId: %d", *id))

	if err := file.Close(); err != nil {
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(err, 500))
		h.metrics.RecordResponse(statusCode, "POST", "CreateNewCourse")
	}

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.NewId(id))
	h.metrics.RecordResponse(statusCode, "POST", "CreateNewCourse")
}

// @Summary Создать модуль
// @Accept json
// @Produce json
// @Description Используется для создания нового модуля в курсе. Требуется токен администратора.
// @Success 200 {object} entity.Id
// @Router /v1/billing/management/createModule [post]
// @Tags Методы взаимодействия с контентом
// @Param module body entity.Module true "Данные модуля"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 409 {object} courseerror.CourseError "Модуль с таким названием или позицией уже существует"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) CreateNewModule(ctx *gin.Context) {
	var statusCode int

	module := entity.NewModule()
	if err := ctx.ShouldBindJSON(&module); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать тело запроса", "CreateNewModule", err.Error(), 10101)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.metrics.RecordResponse(statusCode, "POST", "CreateNewModule")
		return
	}

	id, err := h.contentManagementService.AddModule(ctx, module)
	if err != nil {
		h.logger.Error("не получилось добавить модуль", "CreateNewModule", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "CreateNewModule")
			return
		}
		if err.Code == 13001 {
			statusCode = http.StatusConflict
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "CreateNewModule")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "POST", "CreateNewModule")
		return
	}

	h.logger.Info(fmt.Sprintf("модуль был успешно добавлен админом с ID: %d", ctx.Value("AdminId").(uint)), "CreateNewModule", fmt.Sprintf("moduleId: %d", *id))

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.NewId(id))
	h.metrics.RecordResponse(statusCode, "POST", "CreateNewModule")
}

// @Summary Создать урок
// @Accept mpfd
// @Produce json
// @Description Используется для загрузки нового урока в модуль. Требуется токен администратора.
// @Success 200 {object} entity.Id
// @Router /v1/billing/management/uploadLesson [post]
// @Tags Методы взаимодействия с контентом
// @Param name formData string true "Название урока"
// @Param moduleName formData string true "Название урока"
// @Param description formData string true "Описание урока"
// @Param position formData int true "Позиция урока"
// @Param courseName formData string true "Название курса"
// @Param preview formData file true "Превью"
// @Param lesson formData file true "Урок"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или неверно передано превью фото или урок"
// @Failure 403 {object} courseerror.CourseError "Ошибка авторизации в CDN"
// @Failure 409 {object} courseerror.CourseError "Урок с таким названием или позицией уже существует"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
// @Failure 503 {object} courseerror.CourseError "CDN недоступен"
func (h Handlers) UploadNewLesson(ctx *gin.Context) {
	var statusCode int

	lesson, err := ctx.FormFile("lesson")
	if err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать видео", "UploadNewLesson", err.Error(), 400)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(err, 400))
		h.metrics.RecordResponse(statusCode, "POST", "UploadNewLesson")
		return
	}

	preview, previewHeader, err := ctx.Request.FormFile("preview")
	if err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать фото", "UploadNewLesson", err.Error(), 400)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBadFormData, 400))
		h.metrics.RecordResponse(statusCode, "POST", "UploadNewLesson")
		return
	}

	name := ctx.PostForm("name")
	moduleName := ctx.PostForm("moduleName")
	description := ctx.PostForm("description")
	position := ctx.PostForm("position")
	courseName := ctx.PostForm("courseName")

	lessonId, courseErr := h.contentManagementService.AddLesson(ctx, lesson, name, moduleName, description, position, courseName, previewHeader, &preview)
	if courseErr != nil {
		h.logger.Error("не получилось добавить урок", "UploadNewLesson", courseErr.Message, courseErr.Code)
		if courseErr.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, courseErr)
			h.metrics.RecordResponse(statusCode, "POST", "UploadNewLesson")
			return
		}
		if courseErr.Code == 11050 {
			statusCode = http.StatusForbidden
			ctx.AbortWithStatusJSON(statusCode, courseErr)
			h.metrics.RecordResponse(statusCode, "POST", "UploadNewLesson")
			return
		}
		if courseErr.Code == 13001 || courseErr.Code == 13002 {
			statusCode = http.StatusConflict
			ctx.AbortWithStatusJSON(statusCode, courseErr)
			h.metrics.RecordResponse(statusCode, "POST", "UploadNewLesson")
			return
		}
		if courseErr.Code == 11051 || courseErr.Code == 14002 || courseErr.Code == 11041 {
			statusCode = http.StatusServiceUnavailable
			ctx.AbortWithStatusJSON(statusCode, courseErr)
			h.metrics.RecordResponse(statusCode, "POST", "UploadNewLesson")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, courseErr)
		h.metrics.RecordResponse(statusCode, "POST", "UploadNewLesson")
		return
	}

	h.logger.Info(fmt.Sprintf("урок был успешно добавлен админом с ID: %d", ctx.Value("AdminId").(uint)), "UploadNewLesson", fmt.Sprintf("lessonId: %d", *lessonId))

	if err := preview.Close(); err != nil {
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(err, http.StatusInternalServerError))
		h.metrics.RecordResponse(statusCode, "POST", "UploadNewLesson")
	}

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.NewId(lessonId))
	h.metrics.RecordResponse(statusCode, "POST", "UploadNewLesson")
}

// @Summary Обновить курс
// @Accept mpfd
// @Produce json
// @Description Используется для редактирования курса. Требуется токен администратора.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/editCourse [patch]
// @Tags Методы взаимодействия с контентом
// @Param name formData string true "Название курса"
// @Param description formData string true "Описание курса"
// @Param cost formData int true "Стоимость курса"
// @Param discount formData int false "Скидка"
// @Param preview formData file true "Превью"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или неверно передано превью фото"
// @Failure 403 {object} courseerror.CourseError "Ошибка авторизации в CDN"
// @Failure 404 {object} courseerror.CourseError "Курс не найден"
// @Failure 409 {object} courseerror.CourseError "Курс с таким названием уже существует"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
// @Failure 503 {object} courseerror.CourseError "CDN недоступен"
func (h Handlers) UpdateCourse(ctx *gin.Context) {
	var statusCode int

	var fileNotExists bool
	name := ctx.PostForm("name")
	description := ctx.PostForm("description")
	cost := ctx.PostForm("cost")
	discount := ctx.PostForm("discount")
	courseId := ctx.PostForm("courseId")
	file, header, err := ctx.Request.FormFile("preview")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			fileNotExists = true
		} else {
			statusCode = http.StatusBadRequest
			h.logger.Error("не получилось обработать фото", "UpdateCourse", err.Error(), 400)
			ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBadFormData, 400))
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateCourse")
			return
		}
	}

	if err := h.contentManagementService.ManageCourse(ctx, courseId, name, description, cost, discount, header, &file, fileNotExists); err != nil {
		h.logger.Error("не получилось обновить курс", "UpdateCourse", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateCourse")
			return
		}
		if err.Code == 13003 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateCourse")
			return
		}
		if err.Code == 13001 {
			statusCode = http.StatusConflict
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateCourse")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "PATCH", "UpdateCourse")
		return
	}

	h.logger.Info(fmt.Sprintf("курс был успешно обновлен админом с ID: %d", ctx.Value("AdminId").(uint)), "UpdateCourse", fmt.Sprintf("name: %v", name))

	if err := file.Close(); err != nil {
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(err, 500))
		h.metrics.RecordResponse(statusCode, "POST", "UpdateCourse")
	}

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("данные о курсе успешно отредактированы"))
	h.metrics.RecordResponse(statusCode, "PATCH", "UpdateCourse")
}

// @Summary Обновить модуль
// @Accept json
// @Produce json
// @Description Используется для редактирования модуля. Требуется токен администратора.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/editModule [post]
// @Tags Методы взаимодействия с контентом
// @Param module body entity.Module true "Данные модуля"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 404 {object} courseerror.CourseError "Модуль не найден"
// @Failure 409 {object} courseerror.CourseError "Модуль с таким названием или позицией уже существует"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) UpdateModule(ctx *gin.Context) {
	var statusCode int

	module := entity.NewModule()
	if err := ctx.ShouldBindJSON(&module); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать тело запроса", "UpdateModule", err.Error(), 10101)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.metrics.RecordResponse(statusCode, "PATCH", "UpdateModule")
		return
	}

	if err := h.contentManagementService.ManageModule(ctx, module); err != nil {
		h.logger.Error("не получилось обновить модуль", "UpdateModule", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateModule")
			return
		}
		if err.Code == 13001 {
			statusCode = http.StatusConflict
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateModule")
			return
		}
		if err.Code == 13002 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateModule")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "PATCH", "UpdateModule")
		return
	}

	h.logger.Info(fmt.Sprintf("модуль был успешно обновлен админом с ID: %d", ctx.Value("AdminId").(uint)), "UpdateModule", fmt.Sprintf("name: %v", module.Name))

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("данные о модуле успешно отредактированы"))
	h.metrics.RecordResponse(statusCode, "PATCH", "UpdateModule")
}

// @Summary Обновить урок
// @Accept mpfd
// @Produce json
// @Description Используется для редактирования урока. Требуется токен администратора.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/editLesson [patch]
// @Tags Методы взаимодействия с контентом
// @Param name formData string true "Название урока"
// @Param moduleName formData string true "Название урока"
// @Param description formData string true "Описание урока"
// @Param position formData int true "Позиция урока"
// @Param courseName formData string true "Название курса"
// @Param preview formData file true "Превью"
// @Param lesson formData file true "Урок"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или неверно передано превью фото или урок"
// @Failure 403 {object} courseerror.CourseError "Ошибка авторизации в CDN"
// @Failure 404 {object} courseerror.CourseError "Урок не найден"
// @Failure 409 {object} courseerror.CourseError "Урок с таким названием или позицией уже существует"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
// @Failure 503 {object} courseerror.CourseError "CDN недоступен"
func (h Handlers) UpdateLesson(ctx *gin.Context) {
	var (
		videoNotExists   bool
		previewNotExists bool
		statusCode       int
	)
	lesson, err := ctx.FormFile("lesson")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			videoNotExists = true
		} else {
			statusCode = http.StatusBadRequest
			h.logger.Error("не получилось обработать видео", "UpdateLesson", err.Error(), 400)
			ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBadFormData, 400))
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateLesson")
			return
		}
	}

	preview, previewHeader, err := ctx.Request.FormFile("preview")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			previewNotExists = true
		} else {
			statusCode = http.StatusBadRequest
			h.logger.Error("не получилось обработать фото", "UpdateLesson", err.Error(), 400)
			ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBadFormData, 400))
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateLesson")
			return
		}

	}

	name := ctx.PostForm("name")
	description := ctx.PostForm("description")
	position := ctx.PostForm("position")
	lessonId := ctx.PostForm("lessonId")

	if err := h.contentManagementService.ManageLesson(
		ctx, lesson, name, description, position, lessonId,
		previewHeader, &preview, videoNotExists, previewNotExists); err != nil {
		h.logger.Error("не получилось обновить урок", "UpdateLesson", err.Message, err.Code)
		if err.Code == 13005 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateLesson")
			return
		}
		if err.Code == 11050 {
			statusCode = http.StatusForbidden
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateLesson")
			return
		}
		if err.Code == 13001 {
			statusCode = http.StatusConflict
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateLesson")
			return
		}
		if err.Code == 13005 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateLesson")
			return
		}
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "UpdateLesson")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "PATCH", "UpdateLesson")
	}

	h.logger.Info(fmt.Sprintf("урок был успешно обновлен админом с ID: %d", ctx.Value("AdminId").(uint)), "UpdateLesson", fmt.Sprintf("name: %v", name))

	if err := preview.Close(); err != nil {
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(err, 500))
		h.metrics.RecordResponse(statusCode, "POST", "UpdateLesson")
	}

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("урок успешно отредактирован"))
	h.metrics.RecordResponse(statusCode, "PATCH", "UpdateLesson")
}

// @Summary Изменить видимость курса
// @Produce json
// @Description Используется для изменения видимости курса для пользователей. Не влияет на уже купленные курсы, они будут видимы в профиле. Требуется токен администратора.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/editVisibility [patch]
// @Tags Методы взаимодействия с контентом
// @Param id query string true "Изменить видимость"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 404 {object} courseerror.CourseError "Курс не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ManageVisibility(ctx *gin.Context) {
	var statusCode int

	id := ctx.Query("id")
	if err := h.contentManagementService.ManageShowStatus(ctx, id); err != nil {
		h.logger.Error("не получилось обновить видимость курса", "ManageVisibility", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "ManageVisibility")
			return
		}
		if err.Code == 13003 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "ManageVisibility")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "PATCH", "ManageVisibility")
		return
	}

	h.logger.Info(fmt.Sprintf("видимость курса была успешно обновлена админом с ID: %d", ctx.Value("AdminId").(uint)), "ManageVisibility", fmt.Sprintf("ID курса: %v", id))

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("видимость модуля успешно изменена"))
	h.metrics.RecordResponse(statusCode, "PATCH", "ManageVisibility")
}

// @Summary Удалить модуль
// @Produce json
// @Description Используется для удаления модуля. Требуется токен администратора.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/deleteModule/{id} [delete]
// @Tags Методы взаимодействия с контентом
// @Param id path string true "ID модуля"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 404 {object} courseerror.CourseError "Модуль не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) EraseModule(ctx *gin.Context) {
	var statusCode int

	id := ctx.Param("id")
	if err := h.contentManagementService.RemoveModule(ctx, id); err != nil {
		h.logger.Error("ошибка при удалении модуля", "EraseModule", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "DELETE", "EraseModule")
			return
		}
		if err.Code == 13002 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "DELETE", "EraseModule")
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "DELETE", "EraseModule")
		return
	}

	h.logger.Info(fmt.Sprintf("модуль был успешно удален админом с ID: %d", ctx.Value("AdminId").(uint)), "EraseModule", fmt.Sprintf("ID модуля: %v", id))

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("модуль и вложенные уроки удалены"))
	h.metrics.RecordResponse(statusCode, "DELETE", "EraseModule")
}

// @Summary Удалить урок
// @Produce json
// @Description Используется для удаления урока. Требуется токен администратора.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/deleteLesson{id} [delete]
// @Tags Методы взаимодействия с контентом
// @Param id path string true "ID урока"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 404 {object} courseerror.CourseError "Модуль не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) EraseLesson(ctx *gin.Context) {
	var statusCode int

	id := ctx.Param("id")
	if err := h.contentManagementService.RemoveLesson(ctx, id); err != nil {
		h.logger.Error("ошибка при удалении уроки", "EraseLesson", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "DELETE", "EraseLesson")
			return
		}
		if err.Code == 13005 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "DELETE", "EraseLesson")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "DELETE", "EraseLesson")
		return
	}

	h.logger.Info(fmt.Sprintf("урок был успешно удален админом с ID: %d", ctx.Value("AdminId").(uint)), "EraseLesson", fmt.Sprintf("ID урока: %v", id))

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("урок успешно удален"))
	h.metrics.RecordResponse(statusCode, "DELETE", "EraseLesson")
}
