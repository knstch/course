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
	errUserIdNotFoundInCtx             = errors.New("почта не найдена в контексте")
	errVerificationStatusNotFoundInCtx = errors.New("статус верификации не найден в контексте")
	errUserEmailIsAlreadyVerified      = errors.New("почта пользователя уже верифицирована")
	errUserNotAuthentificated          = errors.New("пользователь не авторизован")
)

// @Summary Зарегестрироваться пользователю
// @Produce json
// @Accept json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/register [post]
// @Tags Методы для авторизации пользователей
// @Param credentials body entity.Credentials true "учетные данные"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация, или не удалось декодировать сообщение, или почта уже занята"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) SignUp(ctx *gin.Context) {
	credentials := entity.NewCredentials()
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "SignUp", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	token, err := h.authService.Register(ctx, credentials)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось зарегистрировать пользователя с почтой %v", credentials.Email), "SignUp", err.Message, err.Code)
		if err.Code == 11001 || err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.SetCookie("auth", *token, 432000, "/", h.address, true, true)

	h.logger.Info("пользователь успешно зарегистрирован", "SignUp", credentials.Email)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пользователь зарегистрирован"))
}

// @Summary Залогиниться пользователю
// @Produce json
// @Accept json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/login [post]
// @Tags Методы для авторизации пользователей
// @Param credentials body entity.Credentials true "учетные данные"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация, или декодирование сообщения, или почта уже занята"
// @Failure 404 {object} courseerror.CourseError "Пользователь не найден"
// @Failure 405 {object} courseerror.CourseError "Пользователь неактивен"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) SignIn(ctx *gin.Context) {
	credentials := entity.NewCredentials()
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "SignIn", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	token, err := h.authService.LogIn(ctx, credentials)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось залогиниться c почтой %v", credentials.Email), "SignIn", err.Message, err.Code)
		if err.Code == 11002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 11011 {
			ctx.AbortWithStatusJSON(http.StatusMethodNotAllowed, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.SetCookie("auth", *token, 432000, "/", h.address, true, true)

	h.logger.Info(fmt.Sprintf("пользователь успешно залогинен c IP: %v", ctx.ClientIP()), "SignIn", credentials.Email)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("доступ разрешен"))
}

// @Summary Подтвердить почту пользователя
// @Produce json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/email/verification [post]
// @Tags Методы для авторизации пользователей
// @Param confirmCode query string true "Код подтверждения"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация, или декодирование сообщения, или почта пользователя уже подтверждена"
// @Failure 403 {object} courseerror.CourseError "Код подтверждения не совпал"
// @Failure 404 {object} courseerror.CourseError "Код подтверждения не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) Verification(ctx *gin.Context) {
	confirmCode := ctx.Query("confirmCode")

	userId, ok := ctx.Get("userId")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errUserIdNotFoundInCtx, 11005))
		return
	}

	verified, ok := ctx.Get("verified")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errVerificationStatusNotFoundInCtx, 11005))
		return
	}

	if verified.(bool) {
		ctx.JSON(http.StatusBadRequest, courseerror.CreateError(errUserEmailIsAlreadyVerified, 11008))
		return
	}

	token, err := h.authService.VerifyEmail(ctx, confirmCode, userId.(uint))
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось верифицировать почту пользователя c ID = %d", userId), "Verification", err.Message, err.Code)
		if err.Code == 11003 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, err)
			return
		}
		if err.Code == 11004 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.SetCookie("auth", *token, 432000, "/", h.address, true, true)

	h.logger.Info("пользователь успешно верифицирован", "Verification", fmt.Sprintf("userId: %d", userId))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("email верифицирован"))
}

// @Summary Отправить новый код подтверждения
// @Produce json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/email/newConfirmKey [get]
// @Tags Методы для авторизации пользователей
// @Param email query string true "почта для отправки кода подтверждения"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или почта пользователя уже подтверждена"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) SendNewCode(ctx *gin.Context) {
	verified, ok := ctx.Get("verified")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errVerificationStatusNotFoundInCtx, 11005))
		return
	}

	if verified.(bool) {
		ctx.JSON(http.StatusBadRequest, courseerror.CreateError(errUserEmailIsAlreadyVerified, 11008))
		return
	}

	email := ctx.Query("email")

	if err := h.authService.SendNewCofirmationCode(ctx, email); err != nil {
		h.logger.Error(fmt.Sprintf("не получилось отправить код подтверждения на почту юзера с ID = %d", ctx.Value("userId")), "SendNewCode", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("код подтверждения успешно отправлен юзеру", "SendNewCode", fmt.Sprint(ctx.Value("userId")))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("код успешно отправлен"))
}

// @Summary Отправить код для восстановления пароля
// @Produce json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/sendRecoveryCode [get]
// @Tags Методы для авторизации пользователей
// @Param email query string true "почта для восстановления пароля"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) SendRecoverPasswordCode(ctx *gin.Context) {
	email := ctx.Query("email")
	if err := h.authService.SendPasswordRecoverRequest(ctx, email); err != nil {
		h.logger.Error(fmt.Sprintf("не получилось отправить код для восстановления пароля на почту %v", email), "SendRecoverPasswordCode", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("код для восстановления пароля успешно отправлен c IP: %v", ctx.ClientIP()), "SendRecoverPasswordCode", email)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("код для восстановления успешно отправлен"))
}

// @Summary Установить новый пароль
// @Accept json
// @Produce json
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/recoverPassword [post]
// @Tags Методы для авторизации пользователей
// @Param recoverCredentials body entity.PasswordRecoverCredentials true "почта, новый пароль и код"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или код подтверждения неверный"
// @Failure 404 {object} courseerror.CourseError "Код не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) SetNewPassword(ctx *gin.Context) {
	recoverCredentials := entity.NewPasswordRecoverCredentials()
	if err := ctx.ShouldBindJSON(&recoverCredentials); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "SendRecoverPasswordCode", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseerror.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.authService.RecoverPassword(ctx, *recoverCredentials); err != nil {
		h.logger.Error(fmt.Sprintf("не получилось изменить пароль пользователя с email = %v", recoverCredentials.Email), "SetNewPassword", err.Message, err.Code)
		if err.Code == 400 || err.Code == 11003 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 11004 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info("пароль успешно изменен", "SetNewPassword", recoverCredentials.Email)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пароль успешно восстановлен"))
}

func (h Handlers) WithCookieAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Request.Cookie("auth")
		if err != nil {
			h.logger.Error(fmt.Sprintf("отсутствуют куки, вызов с IP: %v", ctx.ClientIP()), "WithCookieAuth", err.Error(), 11009)
			ctx.AbortWithStatusJSON(http.StatusForbidden, courseerror.CreateError(errUserNotAuthentificated, 11009))
			return
		}

		if err := h.authService.ValidateAccessToken(ctx, &cookie.Value); err != nil {
			h.logger.Error("не получилось валидировать токен", "WithCookieAuth", err.Message, err.Code)
			if err.Code == 11006 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		payload, tokenError := h.authService.DecodeToken(ctx, cookie.Value)
		if tokenError != nil {
			h.logger.Error("не получилось декодировать токен", "WithCookieAuth", tokenError.Message, tokenError.Code)
			if tokenError.Code == 11006 || tokenError.Code == 11007 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, tokenError)
			return
		}

		ctx.Set("userId", payload.UserID)
		ctx.Set("verified", payload.Verified)

		h.logger.Info(fmt.Sprintf("пользователь успешно перел по URL: %v c IP: %v", ctx.Request.URL.String(), ctx.ClientIP()), "WithCookieAuth", fmt.Sprint(payload.UserID))

		ctx.Next()
	}
}
