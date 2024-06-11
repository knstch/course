package router

import (
	"github.com/gin-gonic/gin"
	"github.com/knstch/course/internal/app/handlers"
)

func RequestsRouter(h *handlers.Handlers) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("auth")
	auth.POST("/register", h.SignUp)
	auth.POST("/login", h.SignIn)

	email := auth.Group("email")
	email.Use(h.WithCookieAuth())
	email.POST("/verification", h.Verification)

	return router
}
