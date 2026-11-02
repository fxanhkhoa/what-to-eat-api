package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseDishPopularityRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := controllers.NewDishPopularityController()

	group.GET("/trending/", controller.GetTrendingDishes)
	group.GET("/", controller.GetPopularDishes)
	group.GET("/:dishSlug/", controller.GetPopularityMetrics)
	group.POST("/:dishSlug/update/", controller.UpdatePopularityMetrics, aG.AuthGuard, rG.RoleGuard([]string{constants.UPDATE_DISH_POPULARITY}))
}
