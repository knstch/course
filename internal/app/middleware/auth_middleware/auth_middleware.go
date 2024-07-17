package authmiddleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/logger"
	"github.com/knstch/course/internal/app/services/token"
)

var (
	errUserNotAuthentificated = errors.New("пользователь не авторизован")
)

func NewMiddleware(logger logger.Logger, config *config.Config, tokenService *token.TokenService) *Middleware {
	return &Middleware{
		logger,
		config.Secret,
		config.AdminSecret,
		tokenService,
		config.TechMetricsLogin,
		config.TechMetricsPassword,
	}
}

type Middleware struct {
	logger          logger.Logger
	userSecret      string
	adminSecret     string
	tokenService    *token.TokenService
	metricsLogin    string
	metricsPassword string
}

// Claims содержит в себе поля, которые хранятся в JWT.
type UserClaims struct {
	jwt.RegisteredClaims
	Iat      int
	Exp      int
	UserID   uint
	Verified bool
}

// Claims содержит в себе типы данных, которые хранятся в JWT.
type AdminClaims struct {
	jwt.RegisteredClaims
	Iat     int
	Exp     int
	AdminId uint
	Role    string
}

func (m Middleware) WithCookieAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Request.Cookie("auth")
		if err != nil {
			m.logger.Error(fmt.Sprintf("отсутствуют куки, вызов с IP: %v", ctx.ClientIP()), "WithCookieAuth", err.Error(), 11009)
			ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errUserNotAuthentificated, 11009))
			return
		}

		if err := m.tokenService.ValidateAccessToken(ctx, &cookie.Value); err != nil {
			m.logger.Error("не получилось валидировать токен", "WithCookieAuth", err.Message, err.Code)
			if err.Code == 11006 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		payload, tokenError := m.tokenService.DecodeUserToken(ctx, cookie.Value)
		if tokenError != nil {
			m.logger.Error("не получилось декодировать токен", "WithCookieAuth", tokenError.Message, tokenError.Code)
			if tokenError.Code == 11006 || tokenError.Code == 11007 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, tokenError)
			return
		}

		ctx.Set("UserId", payload.UserID)
		ctx.Set("verified", payload.Verified)

		m.logger.Info(fmt.Sprintf("пользователь успешно перел по URL: %v c IP: %v", ctx.Request.URL.String(), ctx.ClientIP()), "WithCookieAuth", fmt.Sprint(payload.UserID))

		ctx.Next()
	}
}

func (m Middleware) WithAdminCookieAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Request.Cookie("admin_auth")
		if err != nil {
			m.logger.Error(fmt.Sprintf("не получилось получить куки, запрос с IP: %v", ctx.ClientIP()), "WithAdminCookieAuth", errUserNotAuthentificated.Error(), 11009)
			ctx.AbortWithStatusJSON(http.StatusForbidden, courseError.CreateError(errUserNotAuthentificated, 11009))
			return
		}

		if err := m.tokenService.ValidateAdminAccessToken(ctx, &cookie.Value); err != nil {
			m.logger.Error("не получилось валидировать токен", "WithAdminCookieAuth", err.Message, err.Code)
			if err.Code == 11006 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		payload, tokenError := m.tokenService.DecodeAdminToken(ctx, cookie.Value)
		if tokenError != nil {
			m.logger.Error("не получилось декодировать токен", "WithAdminCookieAuth", tokenError.Message, tokenError.Code)
			if tokenError.Code == 11006 || tokenError.Code == 11007 {
				ctx.AbortWithStatusJSON(http.StatusForbidden, err)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, tokenError)
			return
		}

		ctx.Set("AdminId", payload.AdminId)
		ctx.Set("Role", payload.Role)

		m.logger.Info(fmt.Sprintf("админ перешел по URL: %v c IP: %v", ctx.Request.URL.String(), ctx.ClientIP()), "WithAdminCookieAuth", fmt.Sprint(payload.AdminId))

		ctx.Next()
	}
}

func (m Middleware) WithMetricsAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, password, hasAuth := ctx.Request.BasicAuth()

		if hasAuth && user == m.metricsLogin && password == m.metricsPassword {
			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(http.StatusForbidden, errUserNotAuthentificated)
		}
	}
}
