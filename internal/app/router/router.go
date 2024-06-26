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
	admin.POST("/login", h.LogIn)
	admin.POST("/verify", h.VerifyAuthentificator)

	management := admin.Group("management")
	management.Use(h.WithAdminCookieAuth())
	management.POST("/register", h.CreateAdmin)
	management.GET("/users", h.FindUsersByFilters)
	management.POST("/ban", h.BanUser)
	management.GET("/user", h.GetUserById)
	management.POST("/createCourse", h.CreateNewCourse)
	management.POST("/createModule", h.CreateNewModule)
	management.POST("/uploadLesson", h.UploadNewLesson)
	management.PATCH("/editCourse", h.UpdateCourse)
	management.PATCH("/editModule", h.UpdateModule)
	management.PATCH("/editLesson", h.UpdateLesson)
	management.PATCH("/editVisibility", h.ManageVisibility)
	management.DELETE("/deleteModule/:id", h.EraseModule)
	management.DELETE("/deleteLesson/:id", h.EraseLesson)
	management.PATCH("/manageBillingHost", h.ManageBillingHost)
	management.PATCH("/manageBillingToken", h.ManageAccessToken)
	management.DELETE("/removeAdmin", h.DeleteAdmin)
	management.PATCH("/changeRole", h.ChangeRole)

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
