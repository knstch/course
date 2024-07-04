package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

func (h Handlers) ManageProfile(ctx *gin.Context) {
	userInfo := entity.NewUserInfo()
	if err := ctx.ShouldBindJSON(&userInfo); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "ManageProfile", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	userId, ok := ctx.Get("userId")
	if !ok {
		h.logger.Error("ошибка при получении userId", "ManageProfile", errUserIdNotFoundInCtx.Error(), 11005)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errUserIdNotFoundInCtx, 11005))
		return
	}

	if err := h.userService.FillProfile(ctx, userInfo, userId.(uint)); err != nil {
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

func (h Handlers) ManagePassword(ctx *gin.Context) {
	passwords := entity.CreateNewPasswords()
	if err := ctx.ShouldBindJSON(&passwords); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "ManagePassword", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.userService.EditPassword(ctx, passwords); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении пароля пользователем с ID: %d", ctx.Value("userId").(uint)), "ManagePassword", err.Message, err.Code)
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

	h.logger.Info("пользователь успешно изменил пароль", "ManagePassword", fmt.Sprintf("ID пользователя: %d, IP: %v", ctx.Value("userId").(uint), ctx.ClientIP()))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пароль успешно изменен"))
}

func (h Handlers) ManageEmail(ctx *gin.Context) {
	email := entity.CreateNewEmail()
	if err := ctx.ShouldBindJSON(&email); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "ManageEmail", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	userId, ok := ctx.Get("userId")
	if !ok {
		h.logger.Error("ошибка при получении userId", "ManageEmail", errUserIdNotFoundInCtx.Error(), 11005)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errUserIdNotFoundInCtx, 11005))
		return
	}

	if err := h.userService.EditEmail(ctx, *email, userId.(uint)); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при изменении почты пользователя с ID: %d, на почту: %v",
			ctx.Value("userId").(uint), email.NewEmail), "ManageEmail", err.Message, err.Code)
		if err.Code == 400 || err.Code == 11002 || err.Code == 11001 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("пользователь с ID: %d успешно изменил почту с IP: %v", userId, ctx.ClientIP()), "ManageEmail", fmt.Sprintf("новая почта: %v", email.NewEmail))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("код успешно отправлен"))
}

func (h Handlers) ConfirmEmailChange(ctx *gin.Context) {
	confirmCode := entity.NewConfirmCodeEntity()
	if err := ctx.ShouldBindJSON(&confirmCode); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "ConfirmEmailChange", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	userId, ok := ctx.Get("userId")
	if !ok {
		h.logger.Error("ошибка при получении userId", "ConfirmEmailChange", errUserIdNotFoundInCtx.Error(), 11005)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errUserIdNotFoundInCtx, 11005))
		return
	}

	if err := h.userService.ConfirmEditEmail(ctx, confirmCode, userId.(uint)); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при подтверждения изменения почты пользователя с ID: %d, с кодом: %d",
			userId, confirmCode.Code), "ConfirmEmailChange", err.Message, err.Code)
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

func (h Handlers) ChangeProfilePhoto(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		h.logger.Error("не получилось обработать фото", "ChangeProfilePhoto", err.Error(), 400)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(err, 400))
		return
	}

	if err := h.userService.AddPhoto(ctx, header, &file); err != nil {
		h.logger.Error("не получилось обновить фото", "ChangeProfilePhoto", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 11050 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("фото пользователя с ID: %d успешно изменено", ctx.Value("userId").(uint)), "ChangeProfilePhoto", "")

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("фото успешно обновлено"))
}

func (h Handlers) GetUser(ctx *gin.Context) {
	user, err := h.userService.GetUserInfo(ctx)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при получении информации о пользователе с ID: %v",
			ctx.Value("userId").(uint)), "GetUser", err.Message, err.Code)
		if err.Code == 11101 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("информация о пользователе успешно получена", "GetUser", fmt.Sprintf("ID: %d", ctx.Value("userId").(uint)))

	ctx.JSON(http.StatusOK, user)
}

func (h Handlers) FreezeProfile(ctx *gin.Context) {
	_, ok := ctx.Get("userId")
	if !ok {
		h.logger.Error("ошибка при получении userId", "FreezeProfile", errUserIdNotFoundInCtx.Error(), 11005)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errUserIdNotFoundInCtx, 11005))
		return
	}

	if err := h.userService.DisableProfile(ctx); err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при заморозке пользователя с ID: %d", ctx.Value("userId").(uint)), "FreezeProfile", err.Message, err.Code)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("профиль пользователя успешно заморожен", "FreezeProfile", fmt.Sprintf("ID: %d", ctx.Value("userId").(uint)))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("профиль успешно заморожен"))
}
