package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type PersonalizedRandomController struct{}

func NewPersonalizedRandomController() *PersonalizedRandomController {
	return &PersonalizedRandomController{}
}

func (c *PersonalizedRandomController) GetPersonalizedRandomDishes(ctx echo.Context) error {
	var dto model.PersonalizedRandomDishDto

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	// Parse limit
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}
	dto.Limit = limit

	// Parse meal categories
	mealCategories := ctx.Request().URL.Query()["mealCategories"]
	if len(mealCategories) > 0 {
		dto.MealCategories = &mealCategories
	}

	// Parse exclude dish slugs
	excludeSlugs := ctx.Request().URL.Query()["excludeDishSlugs"]
	if len(excludeSlugs) > 0 {
		dto.ExcludeDishSlugs = &excludeSlugs
	}

	// Parse boolean flags
	dto.IncludeFavorites = ctx.QueryParam("includeFavorites") == "true"
	dto.IncludeSaved = ctx.QueryParam("includeSaved") == "true"
	dto.IncludeTrending = ctx.QueryParam("includeTrending") == "true"
	dto.ApplyDietaryPreferences = ctx.QueryParam("applyDietaryPreferences") == "true"

	// Parse weights (default to 1.0 if not provided)
	favoriteWeight, err := strconv.ParseFloat(ctx.QueryParam("favoriteWeight"), 64)
	if err != nil || favoriteWeight < 0 || favoriteWeight > 1 {
		favoriteWeight = 1.0
	}
	dto.FavoriteWeight = favoriteWeight

	savedWeight, err := strconv.ParseFloat(ctx.QueryParam("savedWeight"), 64)
	if err != nil || savedWeight < 0 || savedWeight > 1 {
		savedWeight = 1.0
	}
	dto.SavedWeight = savedWeight

	interactionWeight, err := strconv.ParseFloat(ctx.QueryParam("interactionWeight"), 64)
	if err != nil || interactionWeight < 0 || interactionWeight > 1 {
		interactionWeight = 1.0
	}
	dto.InteractionWeight = interactionWeight

	trendingWeight, err := strconv.ParseFloat(ctx.QueryParam("trendingWeight"), 64)
	if err != nil || trendingWeight < 0 || trendingWeight > 1 {
		trendingWeight = 1.0
	}
	dto.TrendingWeight = trendingWeight

	ratingWeight, err := strconv.ParseFloat(ctx.QueryParam("ratingWeight"), 64)
	if err != nil || ratingWeight < 0 || ratingWeight > 1 {
		ratingWeight = 1.0
	}
	dto.RatingWeight = ratingWeight

	// Parse randomness level (0-1, default 0.3 for some randomness)
	randomnessLevel, err := strconv.ParseFloat(ctx.QueryParam("randomnessLevel"), 64)
	if err != nil || randomnessLevel < 0 || randomnessLevel > 1 {
		randomnessLevel = 0.3
	}
	dto.RandomnessLevel = randomnessLevel

	svc := service.NewPersonalizedRandomService()
	results, err := svc.GetPersonalizedRandomDishes(dto)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, results)
}

func (c *PersonalizedRandomController) GetUserPreferenceSummary(ctx echo.Context) error {
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := service.NewPersonalizedRandomService()
	summary, err := svc.GetUserPreferenceSummary(claim.ID)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, summary)
}
