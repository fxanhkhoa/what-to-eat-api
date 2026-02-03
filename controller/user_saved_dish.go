package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type UserSavedDishController struct{}

func NewUserSavedDishController() *UserSavedDishController {
	return &UserSavedDishController{}
}

func (c *UserSavedDishController) SaveDish(ctx echo.Context) error {
	var dto model.CreateUserSavedDishDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserSavedDishService()
	savedDish, err := svc.SaveDish(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, savedDish)
}

func (c *UserSavedDishController) UpdateSavedDish(ctx echo.Context) error {
	id := ctx.Param("id")
	var dto model.UpdateUserSavedDishDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.ID = id
	dto.UserID = claim.ID

	svc := service.NewUserSavedDishService()
	savedDish, err := svc.UpdateSavedDish(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, savedDish)
}

func (c *UserSavedDishController) RemoveSavedDish(ctx echo.Context) error {
	dishSlug := ctx.Param("dishSlug")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	dto := model.DeleteUserSavedDishDto{
		UserID:   claim.ID,
		DishSlug: dishSlug,
	}

	svc := service.NewUserSavedDishService()
	if err := svc.RemoveSavedDish(dto, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Saved dish removed successfully"})
}

func (c *UserSavedDishController) GetUserSavedDishes(ctx echo.Context) error {
	var query model.QueryUserSavedDishDto

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

	tags := ctx.Request().URL.Query()["tags"]
	if len(tags) > 0 {
		query.Tags = &tags
	}

	svc := service.NewUserSavedDishService()
	result, err := svc.GetUserSavedDishes(query)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (c *UserSavedDishController) CheckIsSaved(ctx echo.Context) error {
	dishSlug := ctx.Param("dishSlug")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := service.NewUserSavedDishService()
	isSaved, err := svc.IsSaved(claim.ID, dishSlug)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]bool{"isSaved": isSaved})
}
