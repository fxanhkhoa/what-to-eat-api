package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type DishPopularityController struct{}

func NewDishPopularityController() *DishPopularityController {
	return &DishPopularityController{}
}

func (c *DishPopularityController) GetTrendingDishes(ctx echo.Context) error {
	var dto model.TrendingDishesDto

	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}
	dto.Limit = limit

	period := ctx.QueryParam("period")
	if period == "" {
		period = "week"
	}
	dto.Period = period

	mealCategories := ctx.Request().URL.Query()["mealCategories"]
	if len(mealCategories) > 0 {
		dto.MealCategories = &mealCategories
	}

	svc := service.NewDishPopularityService()
	dishes, err := svc.GetTrendingDishes(dto)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, dishes)
}

func (c *DishPopularityController) GetPopularDishes(ctx echo.Context) error {
	var query model.QueryDishPopularityDto

	var err error
	query.BaseDto.Page, err = strconv.Atoi(ctx.QueryParam("page"))
	if err != nil || query.BaseDto.Page < 0 {
		query.BaseDto.Page = 1
	}

	query.BaseDto.Limit, err = strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || query.BaseDto.Limit < 0 {
		query.BaseDto.Limit = 10
	}

	minAvgRatingStr := ctx.QueryParam("minAverageRating")
	if minAvgRatingStr != "" {
		minAvgRating, err := strconv.ParseFloat(minAvgRatingStr, 64)
		if err == nil {
			query.MinAverageRating = &minAvgRating
		}
	}

	minTrendingScoreStr := ctx.QueryParam("minTrendingScore")
	if minTrendingScoreStr != "" {
		minTrendingScore, err := strconv.ParseFloat(minTrendingScoreStr, 64)
		if err == nil {
			query.MinTrendingScore = &minTrendingScore
		}
	}

	minPopularityScoreStr := ctx.QueryParam("minPopularityScore")
	if minPopularityScoreStr != "" {
		minPopularityScore, err := strconv.ParseFloat(minPopularityScoreStr, 64)
		if err == nil {
			query.MinPopularityScore = &minPopularityScore
		}
	}

	sortBy := ctx.QueryParam("sortBy")
	if sortBy != "" {
		query.SortBy = &sortBy
	}

	svc := service.NewDishPopularityService()
	result, err := svc.GetPopularDishes(query)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (c *DishPopularityController) GetPopularityMetrics(ctx echo.Context) error {
	dishSlug := ctx.Param("dishSlug")

	svc := service.NewDishPopularityService()
	metrics, err := svc.GetPopularityMetrics(dishSlug)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	if metrics == nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"message": "Popularity metrics not found"})
	}

	return ctx.JSON(http.StatusOK, metrics)
}

func (c *DishPopularityController) UpdatePopularityMetrics(ctx echo.Context) error {
	dishSlug := ctx.Param("dishSlug")
	dishID := ctx.QueryParam("dishId")

	svc := service.NewDishPopularityService()
	if err := svc.UpdatePopularityMetrics(dishID, dishSlug); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Popularity metrics updated successfully"})
}
