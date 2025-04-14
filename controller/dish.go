package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/gorilla/mux"
	"github.com/labstack/echo/v4"
)

type DishController struct{}

func NewDishController() *DishController {
	return &DishController{}
}

func (dc *DishController) Find(c echo.Context) {
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
			return
		}
		query.PreparationTimeFrom = &num
	}

	preparationTimeToStr := c.QueryParam("preparationTimeTo")
	if preparationTimeToStr != "" {
		num, err := strconv.Atoi(preparationTimeToStr)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		query.PreparationTimeTo = &num
	}

	cookingTimeFromStr := c.QueryParam("cookingTimeFrom")
	if cookingTimeFromStr == "" {
		num, err := strconv.Atoi(cookingTimeFromStr)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		query.CookingTimeFrom = &num
	}

	cookingTimeToStr := c.QueryParam("cookingTimeTo")
	if cookingTimeToStr != "" {
		num, err := strconv.Atoi(cookingTimeToStr)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
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
		return
	}
	c.JSON(http.StatusOK, helper.NewPaginationHelper().PaginationJson(dishes, count))
}

func (dc *DishController) FindOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dish, err := service.NewDishService().FindOne(vars["slug"])
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	decoded, err := json.Marshal(dish)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(decoded)
}

func (dc *DishController) FindRandom(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}

	limit, errLimit := strconv.Atoi(limitStr)
	if errLimit != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(errLimit.Error()), http.StatusBadRequest)
		return
	}

	dishes, err := service.NewDishService().Random(&limit)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}

	decoded, err := json.Marshal(dishes)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}

	w.Write(decoded)
}
