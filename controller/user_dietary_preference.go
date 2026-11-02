package controller

import (
	"net/http"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type UserDietaryPreferenceController struct{}

func NewUserDietaryPreferenceController() *UserDietaryPreferenceController {
	return &UserDietaryPreferenceController{}
}

func (c *UserDietaryPreferenceController) CreateOrUpdate(ctx echo.Context) error {
	var dto model.CreateUserDietaryPreferenceDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserDietaryPreferenceService()
	preference, err := svc.CreateOrUpdate(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, preference)
}

func (c *UserDietaryPreferenceController) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	var dto model.UpdateUserDietaryPreferenceDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.ID = id
	dto.UserID = claim.ID

	svc := service.NewUserDietaryPreferenceService()
	preference, err := svc.Update(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, preference)
}

func (c *UserDietaryPreferenceController) Get(ctx echo.Context) error {
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := service.NewUserDietaryPreferenceService()
	preference, err := svc.Get(claim.ID)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	if preference == nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"message": "Dietary preferences not found"})
	}

	return ctx.JSON(http.StatusOK, preference)
}

func (c *UserDietaryPreferenceController) Delete(ctx echo.Context) error {
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := service.NewUserDietaryPreferenceService()
	if err := svc.Delete(claim.ID, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Dietary preferences deleted successfully"})
}
