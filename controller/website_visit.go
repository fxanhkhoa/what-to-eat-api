package controller

import (
	"net/http"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type WebsiteVisitController struct{}

func (ctrl *WebsiteVisitController) TrackVisit(c echo.Context) error {
	ip := c.RealIP()
	userAgent := c.Request().UserAgent()
	err := (&service.WebsiteVisitService{}).TrackVisit(ip, userAgent)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusOK)
}

func (ctrl *WebsiteVisitController) CountVisits(c echo.Context) error {
	count, err := (&service.WebsiteVisitService{}).CountVisits()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"count": count})
}
