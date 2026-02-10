package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseUserDietaryPreferenceRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := controllers.NewUserDietaryPreferenceController()

	group.POST("/", controller.CreateOrUpdate, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DIETARY_PREFERENCES}))
	group.GET("/", controller.Get, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DIETARY_PREFERENCES}))
	group.PUT("/:id/", controller.Update, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DIETARY_PREFERENCES}))
	group.DELETE("/", controller.Delete, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DIETARY_PREFERENCES}))
}
