package router

import (
	"what-to-eat/be/controller"

	"github.com/labstack/echo/v4"
)

func UseWebsiteVisitGroup(group *echo.Group) {
	controller := &controller.WebsiteVisitController{}
	group.POST("/visit/", controller.TrackVisit)
	group.GET("/visit/count/", controller.CountVisits)
}
