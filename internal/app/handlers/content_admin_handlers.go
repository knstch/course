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
	name := ctx.PostForm("name")
	description := ctx.PostForm("description")
	cost := ctx.PostForm("cost")
	discount := ctx.PostForm("discount")
	file, header, err := ctx.Request.FormFile("preview")
	if err != nil {
		h.logger.Error("не получилось обработать фото", "CreateNewCourse", err.Error(), 400)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBadFormData, 400))
		return
	}

	id, courseErr := h.contentManagementService.AddCourse(ctx, name, description, cost, discount, header, &file)
	if courseErr != nil {
		h.logger.Error("не получилось добавить курс", "CreateNewCourse", courseErr.Message, courseErr.Code)
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
		if courseErr.Code == 13001 {
			ctx.AbortWithStatusJSON(http.StatusConflict, courseErr)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, courseErr)
		return
	}

	h.logger.Info(fmt.Sprintf("курс был успешно размещен админом с ID: %d", ctx.Value("adminId").(uint)),
		"CreateNewCourse", fmt.Sprintf("courseId: %d", *id))

	ctx.JSON(http.StatusOK, entity.NewId(id))
}

// @Summary Создать модуль
// @Accept json
// @Produce json
// @Success 200 {object} entity.Id
// @Router /v1/billing/management/createModule [post]
// @Tags Методы взаимодействия с контентом
// @Param module body entity.Module true "данные модуля"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 409 {object} courseerror.CourseError "Модуль с таким названием или позицией уже существует"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) CreateNewModule(ctx *gin.Context) {
	module := entity.NewModule()
	if err := ctx.ShouldBindJSON(&module); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "CreateNewModule", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	id, err := h.contentManagementService.AddModule(ctx, module)
	if err != nil {
		h.logger.Error("не получилось добавить модуль", "CreateNewModule", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 13001 {
			ctx.AbortWithStatusJSON(http.StatusConflict, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("модуль был успешно добавлен админом с ID: %d", ctx.Value("adminId").(uint)), "CreateNewModule", fmt.Sprintf("moduleId: %d", *id))

	ctx.JSON(http.StatusOK, entity.NewId(id))
}

// @Summary Создать урок
// @Accept mpfd
// @Produce json
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
	lesson, err := ctx.FormFile("lesson")
	if err != nil {
		h.logger.Error("не получилось обработать видео", "UploadNewLesson", err.Error(), 400)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(err, 400))
		return
	}

	preview, previewHeader, err := ctx.Request.FormFile("preview")
	if err != nil {
		h.logger.Error("не получилось обработать фото", "UploadNewLesson", err.Error(), 400)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBadFormData, 400))
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
			ctx.AbortWithStatusJSON(http.StatusBadRequest, courseErr)
			return
		}
		if courseErr.Code == 11050 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, courseErr)
			return
		}
		if courseErr.Code == 13001 || courseErr.Code == 13002 {
			ctx.AbortWithStatusJSON(http.StatusConflict, courseErr)
			return
		}
		if courseErr.Code == 11051 || courseErr.Code == 14002 || courseErr.Code == 11041 {
			ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, courseErr)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, courseErr)
		return
	}

	h.logger.Info(fmt.Sprintf("урок был успешно добавлен админом с ID: %d", ctx.Value("adminId").(uint)), "UploadNewLesson", fmt.Sprintf("lessonId: %d", *lessonId))

	ctx.JSON(http.StatusOK, entity.NewId(lessonId))
}

// @Summary Обновить курс
// @Accept mpfd
// @Produce json
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
			h.logger.Error("не получилось обработать фото", "UpdateCourse", err.Error(), 400)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBadFormData, 400))
			return
		}
	}

	if err := h.contentManagementService.ManageCourse(ctx, courseId, name, description, cost, discount, header, &file, fileNotExists); err != nil {
		h.logger.Error("не получилось обновить курс", "UpdateCourse", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 13003 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 13001 {
			ctx.AbortWithStatusJSON(http.StatusConflict, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("курс был успешно обновлен админом с ID: %d", ctx.Value("adminId").(uint)), "UpdateCourse", fmt.Sprintf("name: %v", name))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("данные о курсе успешно отредактированы"))
}

// @Summary Обновить модуль
// @Accept json
// @Produce json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/editModule [post]
// @Tags Методы взаимодействия с контентом
// @Param module body entity.Module true "данные модуля"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 404 {object} courseerror.CourseError "Модуль не найден"
// @Failure 409 {object} courseerror.CourseError "Модуль с таким названием или позицией уже существует"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) UpdateModule(ctx *gin.Context) {
	module := entity.NewModule()
	if err := ctx.ShouldBindJSON(&module); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "UpdateModule", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.contentManagementService.ManageModule(ctx, module); err != nil {
		h.logger.Error("не получилось обновить модуль", "UpdateModule", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 13001 {
			ctx.AbortWithStatusJSON(http.StatusConflict, err)
			return
		}
		if err.Code == 13002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("модуль был успешно обновлен админом с ID: %d", ctx.Value("adminId").(uint)), "UpdateModule", fmt.Sprintf("name: %v", module.Name))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("данные о модуле успешно отредактированы"))
}

// @Summary Обновить урок
// @Accept mpfd
// @Produce json
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
	)
	lesson, err := ctx.FormFile("lesson")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			videoNotExists = true
		} else {
			h.logger.Error("не получилось обработать видео", "UpdateLesson", err.Error(), 400)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBadFormData, 400))
			return
		}
	}

	preview, previewHeader, err := ctx.Request.FormFile("preview")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			previewNotExists = true
		} else {
			h.logger.Error("не получилось обработать фото", "UpdateLesson", err.Error(), 400)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBadFormData, 400))
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
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 11050 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, err)
			return
		}
		if err.Code == 13001 {
			ctx.AbortWithStatusJSON(http.StatusConflict, err)
			return
		}
		if err.Code == 13005 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
	}

	h.logger.Info(fmt.Sprintf("урок был успешно обновлен админом с ID: %d", ctx.Value("adminId").(uint)), "UpdateLesson", fmt.Sprintf("name: %v", name))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("урок успешно отредактирован"))
}

// @Summary Изменить видимость курса
// @Produce json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/editVisibility [patch]
// @Tags Методы взаимодействия с контентом
// @Param id query string true "Изменить видимость"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 404 {object} courseerror.CourseError "Курс не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ManageVisibility(ctx *gin.Context) {
	id := ctx.Query("id")
	if err := h.contentManagementService.ManageShowStatus(ctx, id); err != nil {
		h.logger.Error("не получилось обновить видимость курса", "ManageVisibility", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 13003 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("видимость курса была успешно обновлена админом с ID: %d", ctx.Value("adminId").(uint)), "ManageVisibility", fmt.Sprintf("ID курса: %v", id))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("видимость модуля успешно изменена"))
}

// @Summary Удалить модуль
// @Produce json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/deleteModule/{id} [delete]
// @Tags Методы взаимодействия с контентом
// @Param id path string true "ID модуля"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 404 {object} courseerror.CourseError "Модуль не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) EraseModule(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := h.contentManagementService.RemoveModule(ctx, id); err != nil {
		h.logger.Error("ошибка при удалении модуля", "EraseModule", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 13002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("модуль был успешно удален админом с ID: %d", ctx.Value("adminId").(uint)), "EraseModule", fmt.Sprintf("ID модуля: %v", id))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("модуль и вложенные уроки удалены"))
}

// @Summary Удалить урок
// @Produce json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/billing/management/deleteLesson{id} [delete]
// @Tags Методы взаимодействия с контентом
// @Param id path string true "ID урока"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 404 {object} courseerror.CourseError "Модуль не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) EraseLesson(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := h.contentManagementService.RemoveLesson(ctx, id); err != nil {
		h.logger.Error("ошибка при удалении уроки", "EraseLesson", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 13005 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("урок был успешно удален админом с ID: %d", ctx.Value("adminId").(uint)), "EraseLesson", fmt.Sprintf("ID урока: %v", id))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("урок успешно удален"))
}
