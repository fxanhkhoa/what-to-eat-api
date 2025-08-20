package router

import (
	"what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseAuthGroup(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	// rG := middleware.NewRoleGuard()
	controller := &controller.AuthController{}
	group.POST("/login/", controller.Login)
	group.POST("/refresh-token/", controller.RefreshToken)
	group.POST("/logout/", controller.Logout, aG.AuthGuard)
}
