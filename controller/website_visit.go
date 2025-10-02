package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type WebsiteVisitController struct{}

// TrackVisitDto is used to optionally accept IP from frontend
type TrackVisitDto struct {
	IP string `json:"ip,omitempty"`
}

func (ctrl *WebsiteVisitController) TrackVisit(c echo.Context) error {
	// Try to get IP from request body first
	var dto TrackVisitDto
	ip := c.RealIP() // Default to server-detected IP

	// If client sends IP in request body, use that instead
	if err := c.Bind(&dto); err == nil && dto.IP != "" {
		ip = dto.IP
	}

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

func (ctrl *WebsiteVisitController) Find(c echo.Context) error {
	var query model.BaseDto

	var err error
	query.Page, err = strconv.Atoi(c.QueryParam("page"))
	if err != nil || query.Page < 0 {
		query.Page = 1
	}

	query.Limit, err = strconv.Atoi(c.QueryParam("limit"))
	if err != nil || query.Limit < 0 {
		query.Limit = 10
	}

	visits, count, err := (&service.WebsiteVisitService{}).Find(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, helper.PaginationObject{
		Data:  visits,
		Count: count,
	})
}
