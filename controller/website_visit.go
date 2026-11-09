package controller

import (
	"net/http"
	"strconv"
	"strings"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type WebsiteVisitController struct{}

func (ctrl *WebsiteVisitController) TrackVisit(c echo.Context) error {
	var dto model.TrackVisitDto

	// Bind JSON body if present
	_ = c.Bind(&dto)

	ip := c.RealIP()
	if dto.IP == "" {
		dto.IP = ip
	}

	// Read Cloudflare or reverse proxy headers if available
	cfCountry := c.Request().Header.Get("CF-IPCountry")
	xCountry := c.Request().Header.Get("X-Country-Code")
	headerCountryCode := cfCountry
	if headerCountryCode == "" {
		headerCountryCode = xCountry
	}
	headerCountry := c.Request().Header.Get("X-Country-Name")

	userAgent := c.Request().UserAgent()

	err := (&service.WebsiteVisitService{}).TrackVisit(dto, ip, userAgent, headerCountry, headerCountryCode)
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
	var query model.WebsiteVisitQueryDto

	var err error
	query.Page, err = strconv.Atoi(c.QueryParam("page"))
	if err != nil || query.Page < 1 {
		query.Page = 1
	}

	query.Limit, err = strconv.Atoi(c.QueryParam("limit"))
	if err != nil || query.Limit < 1 {
		query.Limit = 10
	}

	query.Country = strings.TrimSpace(c.QueryParam("country"))
	query.Action = strings.TrimSpace(c.QueryParam("action"))
	query.Search = strings.TrimSpace(c.QueryParam("search"))

	if daysStr := c.QueryParam("days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 {
			query.Days = days
		}
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

func (ctrl *WebsiteVisitController) GetAnalytics(c echo.Context) error {
	days := 30
	if daysStr := c.QueryParam("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	analytics, err := (&service.WebsiteVisitService{}).GetAnalytics(days)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, analytics)
}
