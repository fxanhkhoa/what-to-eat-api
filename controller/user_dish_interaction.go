package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type UserDishInteractionController struct{}

func NewUserDishInteractionController() *UserDishInteractionController {
	return &UserDishInteractionController{}
}

func (c *UserDishInteractionController) RecordView(ctx echo.Context) error {
	var dto model.RecordDishViewDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserDishInteractionService()
	if err := svc.RecordView(dto, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "View recorded successfully"})
}

func (c *UserDishInteractionController) RecordCooked(ctx echo.Context) error {
	var dto model.RecordDishCookedDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserDishInteractionService()
	if err := svc.RecordCooked(dto, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Cooked status recorded successfully"})
}

func (c *UserDishInteractionController) RateDish(ctx echo.Context) error {
	var dto model.RateDishDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserDishInteractionService()
	if err := svc.RateDish(dto, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Rating recorded successfully"})
}

func (c *UserDishInteractionController) RecordShare(ctx echo.Context) error {
	var dto model.RecordDishSharedDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserDishInteractionService()
	if err := svc.RecordShare(dto, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Share recorded successfully"})
}

func (c *UserDishInteractionController) GetUserInteractions(ctx echo.Context) error {
	var query model.QueryUserDishInteractionDto

	var err error
	query.BaseDto.Page, err = strconv.Atoi(ctx.QueryParam("page"))
	if err != nil || query.BaseDto.Page < 0 {
		query.BaseDto.Page = 1
	}

	query.BaseDto.Limit, err = strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || query.BaseDto.Limit < 0 {
		query.BaseDto.Limit = 10
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	userID := claim.ID
	query.UserID = &userID

	dishSlug := ctx.QueryParam("dishSlug")
	if dishSlug != "" {
		query.DishSlug = &dishSlug
	}

	cookedStr := ctx.QueryParam("cooked")
	if cookedStr != "" {
		cooked := cookedStr == "true"
		query.Cooked = &cooked
	}

	minRatingStr := ctx.QueryParam("minRating")
	if minRatingStr != "" {
		minRating, err := strconv.Atoi(minRatingStr)
		if err == nil {
			query.MinRating = &minRating
		}
	}

	svc := service.NewUserDishInteractionService()
	result, err := svc.GetUserInteractions(query)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (c *UserDishInteractionController) GetTopInteractedDishes(ctx echo.Context) error {
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	svc := service.NewUserDishInteractionService()
	dishes, err := svc.GetTopInteractedDishes(claim.ID, limit)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, dishes)
}
