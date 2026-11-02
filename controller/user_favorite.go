package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type UserFavoriteController struct{}

func NewUserFavoriteController() *UserFavoriteController {
	return &UserFavoriteController{}
}

func (c *UserFavoriteController) AddFavorite(ctx echo.Context) error {
	var dto model.CreateUserFavoriteDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserFavoriteService()
	favorite, err := svc.AddFavorite(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, favorite)
}

func (c *UserFavoriteController) RemoveFavorite(ctx echo.Context) error {
	dishSlug := ctx.Param("dishSlug")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	dto := model.DeleteUserFavoriteDto{
		UserID:   claim.ID,
		DishSlug: dishSlug,
	}

	svc := service.NewUserFavoriteService()
	if err := svc.RemoveFavorite(dto, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Favorite removed successfully"})
}

func (c *UserFavoriteController) GetUserFavorites(ctx echo.Context) error {
	var query model.QueryUserFavoriteDto

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

	svc := service.NewUserFavoriteService()
	result, err := svc.GetUserFavorites(query)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (c *UserFavoriteController) CheckIsFavorite(ctx echo.Context) error {
	dishSlug := ctx.Param("dishSlug")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := service.NewUserFavoriteService()
	isFavorite, err := svc.IsFavorite(claim.ID, dishSlug)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]bool{"isFavorite": isFavorite})
}
