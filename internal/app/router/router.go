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
	auth.POST("/sendRecoveryCode", h.SendRecoverPasswordCode)
	auth.POST("/recoverPassword", h.SetNewPassword)

	email := auth.Group("email")
	email.Use(h.WithCookieAuth())
	email.POST("/verification", h.Verification)
	email.GET("/newConfirmKey", h.SendNewCode)

	profile := v1.Group("profile")
	profile.Use(h.WithCookieAuth())
	profile.PATCH("/editProfile", h.ManageProfile)
	profile.PATCH("/editPassword", h.ManagePassword)
	profile.POST("/editEmail", h.ManageEmail)
	profile.POST("/confirmEmailChange", h.ConfirmEmailChange)
	profile.POST("/setPhoto", h.ChangeProfilePhoto)
	profile.GET("/getUser", h.GetUser)
	profile.GET("/getCourses", h.RetreiveCourses)

	admin := v1.Group("admin")
	admin.POST("/register", h.CreateAdmin)
	admin.POST("verify", h.VerifyAuthentificator)
	admin.GET("/users", h.FindUsersByFilters)
	admin.POST("/ban", h.BanUser)
	admin.GET("/user", h.GetUserById)
	admin.POST("/createCourse", h.CreateNewCourse)
	admin.POST("/createModule", h.CreateNewModule)
	admin.POST("/uploadLesson", h.UploadNewLesson)
	admin.PATCH("/editCourse", h.UpdateCourse)
	admin.PATCH("/editModule", h.UpdateModule)
	admin.PATCH("/editLesson", h.UpdateLesson)
	admin.PATCH("/editVisibility", h.ManageVisibility)
	admin.DELETE("/deleteModule/:id", h.EraseModule)
	admin.DELETE("/deleteLesson/:id", h.EraseLesson)
	admin.PATCH("/manageBillingHost", h.ManageBillingHost)
	admin.PATCH("/manageBillingToken", h.ManageAccessToken)

	content := v1.Group("content")
	content.GET("/courses", h.RetreiveCourses)
	content.GET("/modules", h.RetreiveModules)
	content.GET("/lessons", h.RetreiveLessons)

	billing := v1.Group("billing")
	billing.Use(h.WithCookieAuth())
	billing.POST("/buyCourse", h.BuyCourse)
	billing.GET("/successPayment/:userData", h.CompletePurchase)
	billing.GET("/failPayment/:userData", h.DeclineOrder)

	return router
}
