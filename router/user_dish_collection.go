package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseUserDishCollectionRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := controllers.NewUserDishCollectionController()

	group.POST("/", controller.Create, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_COLLECTIONS}))
	group.GET("/", controller.GetCollections, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_COLLECTIONS}))
	group.GET("/:id/", controller.GetCollectionByID, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_COLLECTIONS}))
	group.PUT("/:id/", controller.Update, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_COLLECTIONS}))
	group.DELETE("/:id/", controller.Delete, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_COLLECTIONS}))
	group.POST("/add-dish/", controller.AddDish, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_COLLECTIONS}))
	group.DELETE("/:id/dishes/:dishSlug/", controller.RemoveDish, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_COLLECTIONS}))
	group.PUT("/reorder/", controller.ReorderDishes, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_COLLECTIONS}))
	group.POST("/duplicate/", controller.Duplicate, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_COLLECTIONS}))
}
