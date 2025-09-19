package router

import (
	"what-to-eat/be/constants"
	"what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseWebsiteVisitGroup(group *echo.Group) {
	controller := &controller.WebsiteVisitController{}
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()

	group.POST("/visit/", controller.TrackVisit)
	group.GET("/visit/count/", controller.CountVisits, aG.AuthGuard, rG.RoleGuard([]string{constants.FIND_ALL_WEBSITE_VISIT}))
	group.GET("/visit/", controller.Find, aG.AuthGuard, rG.RoleGuard([]string{constants.FIND_ALL_WEBSITE_VISIT}))
}
