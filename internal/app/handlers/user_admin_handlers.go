package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

// @Summary Найти пользователей по фильтрам
// @Produce json
// @Description Используется для поиска пользователей по фильтрам. Требуется токен администратора.
// @Success 200 {object} entity.UserDataWithPagination
// @Router /v1/admin/management/users [get]
// @Tags Методы для администрирования
// @Param firstName query string false "Имя"
// @Param surname query string false "Фамилия"
// @Param phoneNumber query string false "Номер телефона"
// @Param email query string false "Почта"
// @Param active query string false "Активный аккаунт"
// @Param verified query string false "Верифицированный аккаунт"
// @Param courseName query string false "Название курса"
// @Param banned query string false "Статус бана"
// @Param page query string true "Страница"
// @Param limit query string true "Лимит"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) FindUsersByFilters(ctx *gin.Context) {
	firstName := ctx.Query("firstName")
	surname := ctx.Query("surname")
	phoneNumber := ctx.Query("phoneNumber")
	email := ctx.Query("email")
	active := ctx.Query("active")
	verified := ctx.Query("verified")
	courseName := ctx.Query("courseName")
	banned := ctx.Query("banned")
	page := ctx.Query("page")
	limit := ctx.Query("limit")
	users, err := h.userManagementService.RetreiveUsersByFilters(ctx, firstName, surname,
		phoneNumber, email, active, verified, courseName, banned, page, limit)
	if err != nil {
		h.logger.Error(
			fmt.Sprintf("ошибка при поиске по фильтрам: firstName - %v, surname - %v, phoneNumber - %v, email - %v, active - %v, verified - %v, courseName - %v, banned - %v, page - %v, limit- %v", firstName, surname,
				phoneNumber, email, active, verified, courseName, banned, page, limit), "FindUsersByFilters", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("пользователи успешно получены по фильртрам админом с ID: %d",
		ctx.Value("adminId").(uint)),
		"FindUsersByFilters",
		fmt.Sprintf("фильтры: firstName - %v, surname - %v, phoneNumber - %v, email - %v, active - %v, verified - %v, courseName - %v, banned - %v, page - %v, limit- %v",
			firstName, surname, phoneNumber, email, active, verified, courseName, banned, page, limit))

	ctx.JSON(http.StatusOK, users)
}

// @Summary Заблокировать пользователя
// @Produce json
// @Description Используется для блокировки пользователей по ID.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/admin/management/ban [post]
// @Tags Методы для администрирования
// @Param id query string true "ID"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) BanUser(ctx *gin.Context) {
	id := ctx.Query("id")
	if err := h.userManagementService.DeactivateUser(ctx, id); err != nil {
		h.logger.Error("ошибка при блокировке пользователя", "BanUser", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("пользователь успешно забанен админом с ID: %v",
		ctx.Value("adminId").(uint)), "BanUser", fmt.Sprintf("userId: %v", id))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пользователь успешно заблокирован"))
}

// @Summary Разблокировать пользователя
// @Produce json
// @Description Используется для разбана пользователей по ID. Требуется токен администратора.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/admin/management/ban [post]
// @Tags Методы для администрирования
// @Param id query string true "ID"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) UnbanUser(ctx *gin.Context) {
	id := ctx.Query("id")

	if err := h.userManagementService.ActivateUser(ctx, id); err != nil {
		h.logger.Error("ошибка при разблокировке пользователя", "UnbanUser", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("пользователь успешно разблокирован админом с ID: %v",
		ctx.Value("adminId").(uint)), "UnbanUser", fmt.Sprintf("userId: %v", id))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пользователь успешно разблокирован"))
}

// @Summary Получить пользователя по ID
// @Produce json
// @Description Используется для получения данных о пользователе по ID. Требуется токен администратора.
// @Success 200 {object} entity.UserDataAdmin
// @Router /v1/admin/management/user [get]
// @Tags Методы для администрирования
// @Param id query string true "ID"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 404 {object} courseerror.CourseError "Пользователь не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) GetUserById(ctx *gin.Context) {
	id := ctx.Query("id")
	user, err := h.userManagementService.RetreiveUserById(ctx, id)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при получении пользователя по ID: %v", id), "GetUserById", err.Message, err.Code)
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

	h.logger.Info(fmt.Sprintf("пользователь успешно получен админом с ID: %d", ctx.Value("adminId").(uint)), "GetUserById", fmt.Sprintf("userId: %v", id))

	ctx.JSON(http.StatusOK, user)
}

// @Summary Изменить профиль пользовтаеля
// @Accept json
// @Produce json
// @Description Используется для редактирования профиля администратором. Требуется токен администратора.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/admin/management/editUserProfile [patch]
// @Tags Методы для администрирования
// @Param userInfo body entity.UserInfo true "Новые данные"
// @Param id query string true "ID пользователя"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или не удалось декодировать сообщение"
// @Failure 404 {object} courseerror.CourseError "Пользователь не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) EditUserProfile(ctx *gin.Context) {
	id := ctx.Query("id")
	userInfo := entity.NewUserInfo()
	if err := ctx.ShouldBindJSON(&userInfo); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "EditUserProfile", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.userService.FillProfile(ctx, userInfo, id, true); err != nil {
		h.logger.Error("ошибка при заполнении профиля", "EditUserProfile", err.Message, err.Code)
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

	h.logger.Info(fmt.Sprintf("информация профиля пользователя с ID: %v админом с ID: %d успешно изменена", id, ctx.Value("adminId").(uint)), "EditUserProfile", "")

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("данные успешно изменены"))
}

// @Summary Удалить фото пользователя
// @Produce json
// @Description Используется для удаления фото пользователя администратором. Требуется токен администратора.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/admin/management/deleteProfilePhoto [delete]
// @Tags Методы для администрирования
// @Param id query string true "ID"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 404 {object} courseerror.CourseError "Пользователь не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) RemoveUserProfilePhoto(ctx *gin.Context) {
	id := ctx.Query("id")
	if err := h.userManagementService.EraseUserProfilePhoto(ctx, id); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при удаления фото профиля пользователя с ID: %v", id), "RemoveUserProfilePhoto", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 11101 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("фото профиля пользователя с ID: %v админом с ID: %d успешно удалено", id, ctx.Value("adminId").(uint)), "RemoveUserProfilePhoto", "")

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("фото успешно удалено"))
}
