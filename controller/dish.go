package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type DishController struct{}

func NewDishController() *DishController {
	return &DishController{}
}

func (dc *DishController) Find(c echo.Context) error {
	var query model.QueryDishDto

	var err error
	query.BaseDto.Page, err = strconv.Atoi(c.QueryParam("page"))
	if err != nil || query.BaseDto.Page < 0 {
		query.BaseDto.Page = 1
	}

	query.BaseDto.Limit, err = strconv.Atoi(c.QueryParam("limit"))
	if err != nil || query.BaseDto.Limit < 0 {
		query.BaseDto.Limit = 10
	}

	keyword := c.QueryParam("keyword")
	if keyword != "" {
		query.Keyword = &keyword
	}

	tags := c.Request().URL.Query()["tags"]
	if len(tags) > 0 {
		query.Tags = &tags
	}

	preparationTimeFromStr := c.QueryParam("preparationTimeFrom")
	if preparationTimeFromStr != "" {
		num, err := strconv.Atoi(preparationTimeFromStr)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		query.PreparationTimeFrom = &num
	}

	preparationTimeToStr := c.QueryParam("preparationTimeTo")
	if preparationTimeToStr != "" {
		num, err := strconv.Atoi(preparationTimeToStr)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		query.PreparationTimeTo = &num
	}

	cookingTimeFromStr := c.QueryParam("cookingTimeFrom")
	if cookingTimeFromStr != "" {
		num, err := strconv.Atoi(cookingTimeFromStr)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		query.CookingTimeFrom = &num
	}

	cookingTimeToStr := c.QueryParam("cookingTimeTo")
	if cookingTimeToStr != "" {
		num, err := strconv.Atoi(cookingTimeToStr)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		query.CookingTimeTo = &num
	}

	q := c.Request().URL.Query()

	difficultLevels := q["difficultLevels"]
	query.DifficultLevels = &difficultLevels

	mealCategories := q["mealCategories"]
	query.MealCategories = &mealCategories

	ingredientCategories := q["ingredientCategories"]
	query.IngredientCategories = &ingredientCategories

	ingredients := q["ingredients"]
	query.Ingredients = &ingredients

	labels := q["labels"]
	query.Labels = &labels

	dishService := &service.DishService{}
	dishes, count, err := dishService.Find(query)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, helper.PaginationObject{
		Data:  dishes,
		Count: count,
	})
}

func (dc *DishController) FindOne(c echo.Context) error {
	id := c.Param("id")
	s := &service.DishService{}
	dish, err := s.FindOne(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	return c.JSON(http.StatusOK, dish)
}

func (dc *DishController) FindOneBySlug(c echo.Context) error {
	slug := c.Param("slug")
	s := &service.DishService{}
	dish, err := s.FindOneBySlug(slug)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	return c.JSON(http.StatusOK, dish)
}

func (dc *DishController) FindRandom(c echo.Context) error {
	var query model.QueryDishRandomDto
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || query.Limit < 0 {
		query.Limit = 10
	} else {
		query.Limit = limit
	}

	q := c.Request().URL.Query()

	mealCategories := q["mealCategories"]
	query.MealCategories = &mealCategories

	s := &service.DishService{}
	dishes, err := s.Random(query)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, dishes)
}

func (dc *DishController) Create(c echo.Context) error {
	var dto model.CreateDishDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	var service = &service.DishService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Create(dto, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (dc *DishController) Update(c echo.Context) error {
	id := c.Param("id")
	var dto model.UpdateDishDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	dto.ID = id
	var service = &service.DishService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Update(dto, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (dc *DishController) Remove(c echo.Context) error {
	id := c.Param("id")
	var service = &service.DishService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Remove(id, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}
