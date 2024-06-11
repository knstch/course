package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

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

func (h *Handlers) WithCookieAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Request.Cookie("auth")
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(err, 403))
			return
		}

		if err := h.authService.ValidateAccessToken(ctx, &cookie.Value); err != nil {
			if err.Code == 11006 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		payload, tokenError := h.authService.DecodeToken(ctx, cookie.Value)
		if tokenError != nil {
			if tokenError.Code == 11006 || tokenError.Code == 11007 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		ctx.Set("userId", payload.UserID)
		ctx.Set("verified", payload.Verified)
		ctx.Set("subscriptionType", payload.SubscriptionType)

		ctx.Next()
	}
}
