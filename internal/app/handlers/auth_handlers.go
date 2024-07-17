package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

const (
	fiveDaysInSeconds = 432000
)

var (
	errUserEmailIsAlreadyVerified = errors.New("почта пользователя уже верифицирована")
)

// @Summary Зарегестрироваться пользователю
// @Produce json
// @Accept json
// @Description Используется для регистрации новых пользователей. После регистрации необходимо подтвердить почту.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/register [post]
// @Tags Методы для авторизации пользователей
// @Param credentials body entity.Credentials true "Учетные данные"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация, или не удалось декодировать сообщение, или почта уже занята"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) SignUp(ctx *gin.Context) {
	var statusCode int

	credentials := entity.NewCredentials()
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать тело запроса", "SignUp", err.Error(), 10101)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.metrics.RecordResponse(statusCode, "POST", "SignUp")
		return
	}

	token, err := h.authService.Register(ctx, credentials)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось зарегистрировать пользователя с почтой %v", credentials.Email), "SignUp", err.Message, err.Code)
		if err.Code == 11001 || err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "SignUp")
			return
		}
		if err.Code == 17002 {
			statusCode = http.StatusTooManyRequests
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "SignUp")
			return
		}
		if err.Code == 17003 || err.Code == 17004 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "SignUp")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "POST", "SignUp")
		return
	}

	ctx.SetCookie("auth", *token, fiveDaysInSeconds, "/", h.address, true, true)

	h.logger.Info("пользователь успешно зарегистрирован", "SignUp", credentials.Email)

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("пользователь зарегистрирован"))
	h.metrics.RecordResponse(statusCode, "POST", "SignUp")
}

// @Summary Залогиниться пользователю
// @Produce json
// @Accept json
// @Description Используется для логина пользователей.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/login [post]
// @Tags Методы для авторизации пользователей
// @Param credentials body entity.Credentials true "Учетные данные"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация, или декодирование сообщения, или почта уже занята"
// @Failure 404 {object} courseerror.CourseError "Пользователь не найден"
// @Failure 405 {object} courseerror.CourseError "Пользователь неактивен"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) SignIn(ctx *gin.Context) {
	var statusCode int

	credentials := entity.NewCredentials()
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать тело запроса", "SignIn", err.Error(), 10101)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.metrics.RecordResponse(statusCode, "POST", "SignIn")
		return
	}

	token, err := h.authService.LogIn(ctx, credentials)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось залогиниться c почтой %v", credentials.Email), "SignIn", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "SignIn")
			return
		}
		if err.Code == 11002 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "SignIn")
			return
		}
		if err.Code == 11011 {
			statusCode = http.StatusMethodNotAllowed
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "SignIn")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "POST", "SignIn")
		return
	}

	ctx.SetCookie("auth", *token, fiveDaysInSeconds, "/", h.address, true, true)

	h.logger.Info(fmt.Sprintf("пользователь успешно залогинен c IP: %v", ctx.ClientIP()), "SignIn", credentials.Email)

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("доступ разрешен"))
	h.metrics.RecordResponse(statusCode, "POST", "SignIn")
}

// @Summary Подтвердить почту пользователя
// @Produce json
// @Description Используется для верификации почты пользователя.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/email/verification [post]
// @Tags Методы для авторизации пользователей
// @Param confirmCode query string true "Код подтверждения"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация, или декодирование сообщения, или почта пользователя уже подтверждена"
// @Failure 403 {object} courseerror.CourseError "Код подтверждения не совпал"
// @Failure 404 {object} courseerror.CourseError "Код подтверждения не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) Verification(ctx *gin.Context) {
	var statusCode int

	confirmCode := ctx.Query("confirmCode")
	userId := ctx.Value("UserId").(uint)
	verified := ctx.Value("verified")

	if verified.(bool) {
		statusCode = http.StatusBadRequest
		ctx.JSON(statusCode, courseerror.CreateError(errUserEmailIsAlreadyVerified, 11008))
		h.metrics.RecordResponse(statusCode, "POST", "Verification")
		return
	}

	token, err := h.authService.VerifyEmail(ctx, confirmCode, userId)
	if err != nil {
		h.logger.Error(fmt.Sprintf("не получилось верифицировать почту пользователя c ID = %d", userId), "Verification", err.Message, err.Code)
		if err.Code == 11003 {
			statusCode = http.StatusForbidden
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "Verification")
			return
		}
		if err.Code == 11004 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "Verification")
			return
		}
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "Verification")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "POST", "Verification")
		return
	}

	ctx.SetCookie("auth", *token, fiveDaysInSeconds, "/", h.address, true, true)

	h.logger.Info("пользователь успешно верифицирован", "Verification", fmt.Sprintf("userId: %d", userId))

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("email верифицирован"))
	h.metrics.RecordResponse(statusCode, "POST", "Verification")
}

// @Summary Отправить новый код подтверждения
// @Produce json
// @Description Используется для отправки нового кода на почту.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/email/newConfirmKey [get]
// @Tags Методы для авторизации пользователей
// @Param email query string true "Почта для отправки кода подтверждения"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или почта пользователя уже подтверждена"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) SendNewCode(ctx *gin.Context) {
	var statusCode int

	verified := ctx.Value("verified")

	if verified.(bool) {
		statusCode = http.StatusBadRequest
		ctx.JSON(statusCode, courseerror.CreateError(errUserEmailIsAlreadyVerified, 11008))
		h.metrics.RecordResponse(statusCode, "GET", "SendNewCode")
		return
	}

	email := ctx.Query("email")
	userId := ctx.Value("UserId").(uint)

	if err := h.authService.SendNewCofirmationCode(ctx, email); err != nil {
		h.logger.Error(fmt.Sprintf("не получилось отправить код подтверждения на почту юзера с ID = %d", userId), "SendNewCode", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "SendNewCode")
			return
		}
		if err.Code == 17002 {
			statusCode = http.StatusTooManyRequests
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "SendNewCode")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "GET", "SendNewCode")
		return
	}

	h.logger.Info("код подтверждения успешно отправлен юзеру", "SendNewCode", fmt.Sprint(userId))

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("код успешно отправлен"))
	h.metrics.RecordResponse(statusCode, "GET", "SendNewCode")
}

// @Summary Отправить код для восстановления пароля
// @Produce json
// @Description Используется для восстановления пароля. Отправляет код на почту.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/sendRecoveryCode [get]
// @Tags Методы для авторизации пользователей
// @Param email query string true "Почта для восстановления пароля"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация"
// @Failure 429 {object} courseerror.CourseError "Слишком много запросов"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) SendRecoverPasswordCode(ctx *gin.Context) {
	var statusCode int

	email := ctx.Query("email")
	if err := h.authService.SendPasswordRecoverRequest(ctx, email); err != nil {
		h.logger.Error(fmt.Sprintf("не получилось отправить код для восстановления пароля на почту %v", email), "SendRecoverPasswordCode", err.Message, err.Code)
		if err.Code == 400 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "SendRecoverPasswordCode")
			return
		}
		if err.Code == 17002 {
			statusCode = http.StatusTooManyRequests
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "GET", "SendRecoverPasswordCode")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "GET", "SendRecoverPasswordCode")
		return
	}

	h.logger.Info(fmt.Sprintf("код для восстановления пароля успешно отправлен c IP: %v", ctx.ClientIP()), "SendRecoverPasswordCode", email)

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("код для восстановления успешно отправлен"))
	h.metrics.RecordResponse(statusCode, "GET", "SendRecoverPasswordCode")
}

// @Summary Установить новый пароль
// @Accept json
// @Produce json
// @Description Используется для установки нового пароля. Для подтверждения нужен код с почты.
// @Success 200 {object} entity.SuccessResponse
// @Router /v1/auth/recoverPassword [post]
// @Tags Методы для авторизации пользователей
// @Param recoverCredentials body entity.PasswordRecoverCredentials true "Почта, новый пароль и код"
// @Failure 400 {object} courseerror.CourseError "Провалена валидация или код подтверждения неверный"
// @Failure 404 {object} courseerror.CourseError "Код не найден"
// @Failure 500 {object} courseerror.CourseError "Возникла внутренняя ошибка"
func (h Handlers) SetNewPassword(ctx *gin.Context) {
	var statusCode int

	recoverCredentials := entity.NewPasswordRecoverCredentials()
	if err := ctx.ShouldBindJSON(&recoverCredentials); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.Error("не получилось обработать тело запроса", "SendRecoverPasswordCode", err.Error(), 10101)
		ctx.AbortWithStatusJSON(statusCode, courseerror.CreateError(errBrokenJSON, 10101))
		h.metrics.RecordResponse(statusCode, "POST", "SetNewPassword")
		return
	}

	if err := h.authService.RecoverPassword(ctx, *recoverCredentials); err != nil {
		h.logger.Error(fmt.Sprintf("не получилось изменить пароль пользователя с email = %v", recoverCredentials.Email), "SetNewPassword", err.Message, err.Code)
		if err.Code == 400 || err.Code == 11003 {
			statusCode = http.StatusBadRequest
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "SetNewPassword")
			return
		}
		if err.Code == 11004 {
			statusCode = http.StatusNotFound
			ctx.AbortWithStatusJSON(statusCode, err)
			h.metrics.RecordResponse(statusCode, "POST", "SetNewPassword")
			return
		}
		statusCode = http.StatusInternalServerError
		ctx.AbortWithStatusJSON(statusCode, err)
		h.metrics.RecordResponse(statusCode, "POST", "SetNewPassword")
		return
	}

	h.logger.Info("пароль успешно изменен", "SetNewPassword", recoverCredentials.Email)

	statusCode = http.StatusOK
	ctx.JSON(statusCode, entity.CreateSuccessResponse("пароль успешно восстановлен"))
	h.metrics.RecordResponse(statusCode, "POST", "SetNewPassword")
}
