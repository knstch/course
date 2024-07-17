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
	errNoRights = errors.New("у вас нет прав")
)

// @Summary Создать профиль нового администратора
// @Accept json
// @Success 200 {object} string "image/png"
// @Description Используется для создания нового администратора. Метод доступен только супер админу.
// @Router /v1/admin/management/register [post]
// @Tags Методы для администрирования
// @Param adminData body entity.AdminCredentials true "Логин, пароль"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 403 {object} courseerror.CourseError "Нет прав"
// @Failure 409 {object} courseerror.CourseError "Невозможно создать админа"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) CreateAdmin(ctx *gin.Context) {
	var statusCode int

	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		statusCode = http.StatusBadRequest
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.logger.Error("не получилось обработать тело запроса", "CreateAdmin", err.Error(), 10101)
		h.metrics.RecordResponse(statusCode, "POST", "CreateAdmin")
		return
	}

	role := ctx.Value("Role").(string)
	if role != "super_admin" {
		statusCode = http.StatusForbidden
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errNoRights, 16004))
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("AdminId")), "CreateAdmin", errNoRights.Error(), 16004)
		h.metrics.RecordResponse(statusCode, "POST", "CreateAdmin")
		return
	}

	qr, err := h.adminService.RegisterAdmin(ctx, credentials)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось зарегистрировать админа с логином: %v", credentials.Login), "CreateAdmin", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "CreateAdmin")
			return
		}
		if err.Code == 16001 {
			statusCode = http.StatusConflict
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "CreateAdmin")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "POST", "CreateAdmin")
		return
	}

	h.logger.Info("администратор успешно создан", "CreateAdmin", credentials.Login)

	statusCode = http.StatusOK
	ctx.Data(statusCode, "image/png", qr)
	h.metrics.RecordResponse(statusCode, "POST", "CreateAdmin")
}

// @Summary Верифицировать аутентификатор
// @Accept json
// @Produce json
// @Description Используется для проверки подключенного аутентификатора у админа.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/admin/verify [post]
// @Tags Методы для администрирования
// @Param adminData body entity.AdminCredentials true "Логин, пароль, код подтверждения"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 403 {object} courseerror.CourseError "Неверный код или пара логин-пароль"
// @Failure 404 {object} courseerror.CourseError "Код не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) VerifyAuthentificator(ctx *gin.Context) {
	var statusCode int

	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать тело запроса", "VerifyAuthentificator", err.Error(), 10101)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.metrics.RecordResponse(statusCode, "POST", "VerifyAuthentificator")
		return
	}

	if err := h.adminService.ApproveTwoStepAuth(ctx, credentials.Login, credentials.Password, credentials.Code); err != nil {
		h.logger.Error(fmt.Sprintf("не получилось привязать двойную аутентификацию для админа с логином: %v", credentials.Login), "VerifyAuthentificator", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "VerifyAuthentificator")
			return
		}
		if err.Code == 16002 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "VerifyAuthentificator")
			return
		}
		if err.Code == 16003 || err.Code == 16052 {
			statusCode = http.StatusForbidden
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "VerifyAuthentificator")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "POST", "VerifyAuthentificator")
		return
	}

	h.logger.Info("двойная аутентификация успешно настроена", "VerifyAuthentificator", credentials.Login)

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("аккаунт успешно верифицирован"))
	h.metrics.RecordResponse(statusCode, "POST", "VerifyAuthentificator")
}

// @Summary Залогиниться администратору
// @Accept json
// @Produce json
// @Description Используется для логина администратора.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/admin/login [post]
// @Tags Методы для администрирования
// @Param adminData body entity.AdminCredentials true "Логин, пароль, код подтверждения"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 403 {object} courseerror.CourseError "Неправильный логин, пароль или код"
// @Failure 404 {object} courseerror.CourseError "Администратор не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) LogIn(ctx *gin.Context) {
	var statusCode int

	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать тело запроса", "LogIn", err.Error(), 10101)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.metrics.RecordResponse(statusCode, "POST", "LogIn")
		return
	}

	token, err := h.adminService.SignIn(ctx, credentials.Login, credentials.Password, credentials.Code)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось войти в аккаунт с логином %v", credentials.Login), "LogIn", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "LogIn")
			return
		}
		if err.Code == 16002 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "LogIn")
			return
		}
		if err.Code == 16003 || err.Code == 16052 {
			statusCode = http.StatusForbidden
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "LogIn")
			return
		}
		statusCode = http.StatusOK
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "POST", "LogIn")
		return
	}

	ctx.SetCookie("admin_auth", *token, fiveDaysInSeconds, "/", h.address, true, true)

	h.logger.Info(fmt.Sprintf("админ успешно вошел c IP: %v", ctx.ClientIP()), "LogIn", credentials.Login)

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("доступ разрешен"))
	h.metrics.RecordResponse(statusCode, "POST", "LogIn")
}

// @Summary Изменить пароль администратора
// @Accept json
// @Produce json
// @Description Используется для смены пароля администратора. Метод доступен только супер админу.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/admin/management/resetPassword [patch]
// @Tags Методы для администрирования
// @Param adminData body entity.AdminCredentials true "Логин, пароль"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или декодирование сообщения"
// @Failure 403 {object} courseerror.CourseError "Неправильный логин, пароль или код"
// @Failure 404 {object} courseerror.CourseError "Администратор не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ChangeAdminPassword(ctx *gin.Context) {
	var statusCode int

	role := ctx.Value("Role").(string)
	if role != "super_admin" {
		statusCode = http.StatusForbidden
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("AdminId")), "ChangeAdminPassword", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errNoRights, 16004))
		h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
		return
	}

	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать тело запроса", "ChangeAdminPassword", err.Error(), 10101)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
		return
	}

	if err := h.adminService.ManageAdminPassword(ctx, credentials); err != nil {
		h.logger.Error(fmt.Sprintf("не получилось изменить пароль админа с логином %v", credentials.Login), "ChangeAdminPassword", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
			return
		}
		if err.Code == 16002 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
		return
	}

	h.logger.Info(fmt.Sprintf("пароль успешно изменен у админа с логином %v", credentials.Login), "изменение пароля", "")

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("пароль успешно изменен"))
	h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
}

// @Summary Изменить ключ администратора
// @Success 200 {object} string "image/png"
// @Description Используется для изменения ключа в аутентификаторе. Метод доступен только супер админу.
// @Router /v1/admin/management/resetKey [patch]
// @Tags Методы для администрирования
// @Param login query string true "Логин"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или не получилось декодировать сообщение"
// @Failure 404 {object} courseerror.CourseError "Администратор не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ChangeAdminAuthKey(ctx *gin.Context) {
	var statusCode int

	role := ctx.Value("Role").(string)
	if role != "super_admin" {
		statusCode = http.StatusForbidden
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("AdminId")), "ChangeAdminAuthKey", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errNoRights, 16004))
		h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
		return
	}
	login := ctx.Query("login")
	qr, err := h.adminService.ManageAdminAuthKey(ctx, login)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось изменить ключ для двойной аутентификации у админом с логином %v", login), "ChangeAdminAuthKey", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
			return
		}
		if err.Code == 16002 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
		return
	}

	h.logger.Info("ключ для двойной аутентификации успешно изменен", "ChangeAdminAuthKey", login)

	statusCode = http.StatusOK
	ctx.Data(statusCode, "image/png", qr)
	h.metrics.RecordResponse(statusCode, "PATCH", "ChangeAdminPassword")
}
