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
	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		h.logger.Error("не получилось обработать тело запроса", "CreateAdmin", err.Error(), 10101)
		return
	}

	role := ctx.Value("role").(string)
	if role != "super_admin" {
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseerror.CreateError(errNoRights, 16004))
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "CreateAdmin", errNoRights.Error(), 16004)
		return
	}

	qr, err := h.adminService.RegisterAdmin(ctx, credentials)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось зарегистрировать админа с логином: %v", credentials.Login), "CreateAdmin", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 16001 {
			ctx.AbortWithStatusJSON(http.StatusConflict, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("администратор успешно создан", "CreateAdmin", credentials.Login)

	ctx.Data(http.StatusOK, "image/png", qr)
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
	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "VerifyAuthentificator", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.adminService.ApproveTwoStepAuth(ctx, credentials.Login, credentials.Password, credentials.Code); err != nil {
		h.logger.Error(fmt.Sprintf("не получилось привязать двойную аутентификацию для админа с логином: %v", credentials.Login), "VerifyAuthentificator", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 16002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 16003 || err.Code == 16052 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("двойная аутентификация успешно настроена", "VerifyAuthentificator", credentials.Login)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("аккаунт успешно верифицирован"))
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
	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "LogIn", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	token, err := h.adminService.SignIn(ctx, credentials.Login, credentials.Password, credentials.Code)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось войти в аккаунт с логином %v", credentials.Login), "LogIn", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 16002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 16003 || err.Code == 16052 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusOK, err)
		return
	}

	ctx.SetCookie("admin_auth", *token, 432000, "/", h.address, true, true)

	h.logger.Info(fmt.Sprintf("админ успешно вошел c IP: %v", ctx.ClientIP()), "LogIn", credentials.Login)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("доступ разрешен"))
}

func (h Handlers) WithAdminCookieAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Request.Cookie("admin_auth")
		if err != nil {
			h.logger.Error(fmt.Sprintf("не получилось получить куки, запрос с IP: %v", ctx.ClientIP()), "WithAdminCookieAuth", errUserNotAuthentificated.Error(), 11009)
			ctx.AbortWithStatusJSON(http.StatusForbidden, courseerror.CreateError(errUserNotAuthentificated, 11009))
			return
		}

		if err := h.adminService.ValidateAdminAccessToken(ctx, &cookie.Value); err != nil {
			h.logger.Error("не получилось валидировать токен", "WithAdminCookieAuth", err.Message, err.Code)
			if err.Code == 11006 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		payload, tokenError := h.adminService.DecodeToken(ctx, cookie.Value)
		if tokenError != nil {
			h.logger.Error("не получилось декодировать токен", "WithAdminCookieAuth", tokenError.Message, tokenError.Code)
			if tokenError.Code == 11006 || tokenError.Code == 11007 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, tokenError)
			return
		}

		ctx.Set("adminId", payload.AdminId)
		ctx.Set("role", payload.Role)

		h.logger.Info(fmt.Sprintf("админ перешел по URL: %v c IP: %v", ctx.Request.URL.String(), ctx.ClientIP()), "WithAdminCookieAuth", fmt.Sprint(payload.AdminId))

		ctx.Next()
	}
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
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "ChangeAdminPassword", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseerror.CreateError(errNoRights, 16004))
		return
	}

	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "ChangeAdminPassword", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.adminService.ManageAdminPassword(ctx, credentials); err != nil {
		h.logger.Error(fmt.Sprintf("не получилось изменить пароль админа с логином %v", credentials.Login), "ChangeAdminPassword", err.Message, err.Code)
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

	h.logger.Info(fmt.Sprintf("пароль успешно изменен у админа с логином %v", credentials.Login), "изменение пароля", "")

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пароль успешно изменен"))
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
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		h.logger.Error(fmt.Sprintf("у админа не хватило прав, id: %d", ctx.Value("adminId")), "ChangeAdminAuthKey", errNoRights.Error(), 16004)
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseerror.CreateError(errNoRights, 16004))
		return
	}
	login := ctx.Query("login")
	qr, err := h.adminService.ManageAdminAuthKey(ctx, login)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось изменить ключ для двойной аутентификации у админом с логином %v", login), "ChangeAdminAuthKey", err.Message, err.Code)
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

	h.logger.Info("ключ для двойной аутентификации успешно изменен", "ChangeAdminAuthKey", login)

	ctx.Data(http.StatusOK, "image/png", qr)
}
