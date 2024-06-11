package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/services/auth"
	"github.com/knstch/course/internal/app/storage"
	"github.com/knstch/course/internal/domain/entity"

	courseError "github.com/knstch/course/internal/app/course_error"
)

type Handlers struct {
	authService auth.AuthService
	address     string
}

func NewHandlers(storage *storage.Storage, config *config.Config, redisClient *redis.Client) *Handlers {
	return &Handlers{
		authService: auth.NewAuthService(storage, config, redisClient),
		address:     config.Address,
	}
}

var (
	errEmailNotFoundInCtx = errors.New("почта не найдена в контексте")
)

func (h *Handlers) SignUp(ctx *gin.Context) {
	credentials := entity.NewCredentials()
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(err, 10101))
		return
	}

	token, err := h.authService.Register(ctx, credentials)
	if err != nil {
		if err.Code == 11001 || err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.SetCookie("auth", *token, 432000, "/", h.address, true, true)

	ctx.Status(http.StatusOK)
}

func (h *Handlers) SignIn(ctx *gin.Context) {
	credentials := entity.NewCredentials()
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(err, 10101))
		return
	}

	token, err := h.authService.LogIn(ctx, credentials)
	if err != nil {
		if err.Code == 11002 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.SetCookie("auth", *token, 432000, "/", h.address, true, true)

	ctx.Status(http.StatusOK)
}

func (h *Handlers) Verification(ctx *gin.Context) {
	confirmCode := entity.NewConfirmCodeEntity()
	if err := ctx.ShouldBindJSON(&confirmCode); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(err, 10101))
		return
	}

	userId, ok := ctx.Get("userId")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errEmailNotFoundInCtx, 11005))
		return
	}

	token, err := h.authService.VerifyEmail(ctx, confirmCode.Code, userId.(uint))
	if err != nil {
		if err.Code == 11003 || err.Code == 11004 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	ctx.SetCookie("auth", *token, 432000, "/", h.address, true, true)

	ctx.JSON(http.StatusOK, map[string]bool{
		"verified": true,
	})
}
