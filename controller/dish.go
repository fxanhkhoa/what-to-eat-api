package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	if cookingTimeFromStr == "" {
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

	difficultLevelStr := c.QueryParam("difficultLevels")
	if difficultLevelStr != "" {
		difficultLevels := strings.Split(difficultLevelStr, ",")
		query.DifficultLevels = &difficultLevels
	}

	mealCategoriesStr := c.QueryParam("mealCategories")
	if mealCategoriesStr != "" {
		mealCategories := strings.Split(mealCategoriesStr, ",")
		query.MealCategories = &mealCategories
	}

	ingredientCategoriesStr := c.QueryParam("ingredientCategories")
	if ingredientCategoriesStr != "" {
		ingredientCategories := strings.Split(ingredientCategoriesStr, ",")
		query.IngredientCategories = &ingredientCategories
	}

	ingredientsStr := c.QueryParam("ingredients")
	if ingredientsStr != "" {
		result := strings.Split(ingredientsStr, ",")
		query.Ingredients = &result
	}

	labelStr := c.QueryParam("labels")
	var labels []string
	if labelStr != "" {
		labels = strings.Split(labelStr, ",")
		query.Labels = &labels
	}

	dishService := &service.DishService{}
	dishes, count, err := dishService.Find(query)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	fmt.Println(dishes)

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

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit < 0 {
		limit = 10
	}

	s := &service.DishService{}
	dishes, err := s.Random(&limit)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, dishes)
}
