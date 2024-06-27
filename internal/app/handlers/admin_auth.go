package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

var (
	errNoRights = errors.New("у вас нет прав")
)

func (h Handlers) CreateAdmin(ctx *gin.Context) {
	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	role := ctx.Value("role").(string)
	if role != "super_admin" {
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	qr, err := h.adminService.RegisterAdmin(ctx, credentials)
	if err != nil {
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

	ctx.Data(http.StatusOK, "image/png", qr)
}

func (h Handlers) VerifyAuthentificator(ctx *gin.Context) {
	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.adminService.ApproveTwoStepAuth(ctx, credentials.Login, credentials.Password, credentials.Code); err != nil {
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

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("аккаунт успешно верифицирован", true))
}

func (h Handlers) LogIn(ctx *gin.Context) {
	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	token, err := h.adminService.SignIn(ctx, credentials.Login, credentials.Password, credentials.Code)
	if err != nil {
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
	}

	ctx.SetCookie("admin_auth", *token, 432000, "/", h.address, true, true)

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("доступ разрешен", true))
}

func (h Handlers) WithAdminCookieAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Request.Cookie("admin_auth")
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errUserNotAuthentificated, 11009))
			return
		}

		if err := h.adminService.ValidateAdminAccessToken(ctx, &cookie.Value); err != nil {
			if err.Code == 11006 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		payload, tokenError := h.adminService.DecodeToken(ctx, cookie.Value)
		if tokenError != nil {
			if tokenError.Code == 11006 || tokenError.Code == 11007 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, tokenError)
			return
		}

		ctx.Set("adminId", payload.AdminId)
		ctx.Set("role", payload.Role)

		ctx.Next()
	}
}

func (h Handlers) ChangeAdminPassword(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	var credentials *entity.AdminCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.adminService.ManageAdminPassword(ctx, credentials); err != nil {
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

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пароль успешно изменен", true))
}

func (h Handlers) ChangeAdminAuthKey(ctx *gin.Context) {
	role := ctx.Value("role").(string)
	if role != "super_admin" {
		ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errNoRights, 16004))
		return
	}

	qr, err := h.adminService.ManageAdminAuthKey(ctx, ctx.Query("login"))
	if err != nil {
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

	ctx.Data(http.StatusOK, "image/png", qr)
}
