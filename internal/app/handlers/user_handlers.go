package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

// @Summary Изменить профиль пользователя
// @Accept json
// @Produce json
// @Description Используется для редактирования профиля пользователем.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/profile/editProfile [patch]
// @Tags Методы для администрирования профиля
// @Param UserInfo body entity.UserInfo true "Данные профиля"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или не удалось декодировать сообщение"
// @Failure 404 {object} courseerror.CourseError "Пользователь не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ManageProfile(ctx *gin.Context) {
	userInfo := entity.NewUserInfo()
	if err := ctx.ShouldBindJSON(&userInfo); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "ManageProfile", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	userId := ctx.Value("UserId").(uint)

	if err := h.userService.FillProfile(ctx, userInfo, fmt.Sprint(userId), false); err != nil {
		h.logger.Error("ошибка при заполнении профиля", "ManageProfile", err.Message, err.Code)
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

	h.logger.Info(fmt.Sprintf("информация профиля успешно изменена пользователем с ID: %d", userId), "ManageProfile", "")

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("данные успешно изменены"))
}

// @Summary Изменить пароль пользователя
// @Accept json
// @Produce json
// @Description Используется для изменения пароля пользователя.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/profile/editPassword [patch]
// @Tags Методы для администрирования профиля
// @Param passwords body entity.Passwords true "Пароли"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или не удалось декодировать сообщение"
// @Failure 404 {object} courseerror.CourseError "Пользователь не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ManagePassword(ctx *gin.Context) {
	passwords := entity.CreateNewPasswords()
	if err := ctx.ShouldBindJSON(&passwords); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "ManagePassword", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	userId := ctx.Value("UserId").(uint)

	if err := h.userService.EditPassword(ctx, passwords); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении пароля пользователем с ID: %d", userId), "ManagePassword", err.Message, err.Code)
		if err.Code == 400 || err.Code == 11102 || err.Code == 11104 || err.Code == 11103 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 11002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("пользователь успешно изменил пароль", "ManagePassword", fmt.Sprintf("ID пользователя: %d, IP: %v", userId, ctx.ClientIP()))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пароль успешно изменен"))
}

// @Summary Изменить почту пользователя
// @Produce json
// @Description Используется для изменения почты пользователя.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/profile/editEmail [patch]
// @Tags Методы для администрирования профиля
// @Param email query string true "Почта"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 404 {object} courseerror.CourseError "Пользователь не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ManageEmail(ctx *gin.Context) {
	email := ctx.Query("email")
	userId := ctx.Value("UserId").(uint)

	if err := h.userService.EditEmail(ctx, email, userId); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении почты пользователя с ID: %d, на почту: %v",
			userId, email), "ManageEmail", err.Message, err.Code)
		if err.Code == 400 || err.Code == 11001 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 11002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 17002 {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("пользователь с ID: %d успешно изменил почту с IP: %v", userId, ctx.ClientIP()), "ManageEmail", fmt.Sprintf("новая почта: %v", email))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("код успешно отправлен"))
}

// @Summary Подтвердить изменении почты пользователя
// @Produce json
// @Description Используется для подтверждения изменения почты пользователя.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/profile/confirmEmailChange [post]
// @Tags Методы для администрирования профиля
// @Param confirmCode query string true "Код подтверждения"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 404 {object} courseerror.CourseError "Пользователь не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ConfirmEmailChange(ctx *gin.Context) {
	confirmCode := ctx.Query("confirmCode")

	userId := ctx.Value("UserId").(uint)

	if err := h.userService.ConfirmEditEmail(ctx, confirmCode, userId); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при подтверждения изменения почты пользователя с ID: %d, с кодом: %v",
			userId, confirmCode), "ConfirmEmailChange", err.Message, err.Code)
		if err.Code == 400 || err.Code == 11002 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 11003 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("изменение почты пользователя c ID: %d успешно завершено c IP: %v", userId, ctx.ClientIP()), "ConfirmEmailChange", "")

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("почта успешно изменена"))
}

// @Summary Изменить фото профиля
// @Accept mpfd
// @Produce json
// @Description Используется для изменения фото профиля пользователя.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/profile/setPhoto [patch]
// @Tags Методы для администрирования профиля
// @Param photo formData file true "Фото"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или не получилось обработать фото"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) ChangeProfilePhoto(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("photo")
	if err != nil {
		h.logger.Error("не получилось обработать фото", "ChangeProfilePhoto", err.Error(), 400)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(err, 400))
		return
	}

	if err := h.userService.AddPhoto(ctx, header, &file); err != nil {
		h.logger.Error("не получилось обновить фото", "ChangeProfilePhoto", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("фото пользователя с ID: %d успешно изменено", ctx.Value("UserId").(uint)), "ChangeProfilePhoto", "")

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("фото успешно обновлено"))
}

// @Summary Получить данные профиля
// @Produce json
// @Description Используется для получения данных профиля пользователя.
// @Success 200 {object} entity.UserData
// @Router /v1/profile/getUser [get]
// @Tags Методы для администрирования профиля
// @Failure 404 {object} courseerror.CourseError "Пользователь не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) GetUser(ctx *gin.Context) {
	userId := ctx.Value("UserId").(uint)
	user, err := h.userService.GetUserInfo(ctx)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при получении информации о пользователе с ID: %v",
			userId), "GetUser", err.Message, err.Code)
		if err.Code == 11101 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("информация о пользователе успешно получена", "GetUser", fmt.Sprintf("ID: %d", userId))

	ctx.JSON(http.StatusOK, user)
}

// @Summary Заморозить профиль
// @Produce json
// @Description Используется для заморозки профиля пользователем.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/profile/disable [post]
// @Tags Методы для администрирования профиля
// @Failure 400 {object} courseerror.CourseError "Ошибка получения userId"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) FreezeProfile(ctx *gin.Context) {
	userId := ctx.Value("UserId").(uint)

	if err := h.userService.DisableProfile(ctx); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при заморозке пользователя с ID: %d", userId), "FreezeProfile", err.Message, err.Code)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("профиль пользователя успешно заморожен", "FreezeProfile", fmt.Sprintf("ID: %d", userId))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("профиль успешно заморожен"))
}

// @Summary Пометить урок как пройденный
// @Produce json
// @Description Используется для добавления урока в просмотренный.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/profile/watchLesson [post]
// @Tags Методы для администрирования профиля
// @Failure 404 {object} courseerror.CourseError "Урок не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) WatchVideo(ctx *gin.Context) {
	lessonId := ctx.Query("id")

	if err := h.userService.MarkLessonAsWatched(ctx, lessonId); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при создании статуса просмотра урока с ID: %v", lessonId), "WatchVideo", err.Message, err.Code)
		if err.Code == 13005 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("урок успешно помечен как просмотренный"))
}
