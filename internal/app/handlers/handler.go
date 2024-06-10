package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/services/auth"
	"github.com/knstch/course/internal/app/storage"
	"github.com/knstch/course/internal/domain/entity"

	courseError "github.com/knstch/course/internal/app/course_error"
)

type Handlers struct {
	authService auth.AuthService
}

func NewHandlers(storage *storage.Storage, config *config.Config) *Handlers {
	return &Handlers{
		authService: auth.NewAuthService(storage, config),
	}
}

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

	ctx.SetCookie("auth", *token, 432000, "/", "localhost", true, true)

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

	ctx.SetCookie("auth", *token, 432000, "/", "localhost", true, true)

	ctx.Status(http.StatusOK)
}
