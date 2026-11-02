package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UsePersonalizedRandomRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := controllers.NewPersonalizedRandomController()

	group.GET("/", controller.GetPersonalizedRandomDishes, aG.AuthGuard, rG.RoleGuard([]string{constants.ACCESS_PERSONALIZED_RANDOM}))
	group.GET("/preference-summary/", controller.GetUserPreferenceSummary, aG.AuthGuard, rG.RoleGuard([]string{constants.VIEW_USER_PREFERENCE_SUMMARY}))
}
