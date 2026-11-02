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

	group.POST("/", controller.Create, aG.AuthGuard, rG.RoleGuard([]string{constants.CREATE_USER_DISH_COLLECTION}))
	group.GET("/", controller.GetCollections, aG.AuthGuard, rG.RoleGuard([]string{constants.FIND_ALL_USER_DISH_COLLECTION}))
	group.GET("/:id/", controller.GetCollectionByID, aG.AuthGuard, rG.RoleGuard([]string{constants.FIND_ONE_USER_DISH_COLLECTION}))
	group.PUT("/:id/", controller.Update, aG.AuthGuard, rG.RoleGuard([]string{constants.UPDATE_USER_DISH_COLLECTION}))
	group.DELETE("/:id/", controller.Delete, aG.AuthGuard, rG.RoleGuard([]string{constants.REMOVE_USER_DISH_COLLECTION}))
	group.POST("/add-dish/", controller.AddDish, aG.AuthGuard, rG.RoleGuard([]string{constants.UPDATE_USER_DISH_COLLECTION}))
	group.DELETE("/:id/dishes/:dishSlug/", controller.RemoveDish, aG.AuthGuard, rG.RoleGuard([]string{constants.UPDATE_USER_DISH_COLLECTION}))
	group.PUT("/reorder/", controller.ReorderDishes, aG.AuthGuard, rG.RoleGuard([]string{constants.UPDATE_USER_DISH_COLLECTION}))
	group.POST("/duplicate/", controller.Duplicate, aG.AuthGuard, rG.RoleGuard([]string{constants.CREATE_USER_DISH_COLLECTION}))
}
