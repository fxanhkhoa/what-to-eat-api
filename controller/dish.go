package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"what-to-eat/be/graph/service"
	"what-to-eat/be/helper"

	"github.com/gorilla/mux"
	"github.com/labstack/echo/v4"
)

type DishController struct{}

func NewDishController() *DishController {
	return &DishController{}
}

func (dc *DishController) Find(c echo.Context) {
	var query dto.QueryMemberDto

	var err error
	query.BaseDto.Page, err = strconv.Atoi(c.QueryParam("page"))
	if err != nil || query.BaseDto.Page < 0 {
		query.BaseDto.Page = 1
	}

	query.BaseDto.Limit, err = strconv.Atoi(c.QueryParam("limit"))
	if err != nil || query.BaseDto.Limit < 0 {
		query.BaseDto.Limit = 10
	}

	keywordStr := r.URL.Query().Get("keyword")
	var keyword *string
	if keywordStr == "" {
		keyword = nil
	} else {
		keyword = &keywordStr
	}

	tagsStr := r.URL.Query().Get("tags")
	var tags []string
	if tagsStr == "" {
		tags = []string{}
	} else {
		tags = strings.Split(tagsStr, ",")
	}

	preparationTimeFromStr := r.URL.Query().Get("preparationTimeFrom")
	var preparationTimeFrom *int
	if preparationTimeFromStr == "" {
		preparationTimeFrom = nil
	} else {
		num, err := strconv.Atoi(preparationTimeFromStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		preparationTimeFrom = &num
	}

	preparationTimeToStr := r.URL.Query().Get("preparationTimeTo")
	var preparationTimeTo *int
	if preparationTimeToStr == "" {
		preparationTimeTo = nil
	} else {
		num, err := strconv.Atoi(preparationTimeToStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		preparationTimeTo = &num
	}

	cookingTimeFromStr := r.URL.Query().Get("cookingTimeFrom")
	var cookingTimeFrom *int
	if cookingTimeFromStr == "" {
		cookingTimeFrom = nil
	} else {
		num, err := strconv.Atoi(cookingTimeFromStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		cookingTimeFrom = &num
	}

	cookingTimeToStr := r.URL.Query().Get("cookingTimeTo")
	var cookingTimeTo *int
	if cookingTimeToStr == "" {
		cookingTimeTo = nil
	} else {
		num, err := strconv.Atoi(cookingTimeToStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		cookingTimeTo = &num
	}

	difficultLevelStr := r.URL.Query().Get("difficultLevels")
	var difficultLevels []string
	if difficultLevelStr == "" {
		difficultLevels = []string{}
	} else {
		difficultLevels = strings.Split(difficultLevelStr, ",")
	}

	mealCategoriesStr := r.URL.Query().Get("mealCategories")
	var mealCategories []string
	if mealCategoriesStr == "" {
		mealCategories = []string{}
	} else {
		mealCategories = strings.Split(mealCategoriesStr, ",")
	}

	ingredientCategoriesStr := r.URL.Query().Get("ingredientCategories")
	var ingredientCategories []string
	if ingredientCategoriesStr == "" {
		ingredientCategories = []string{}
	} else {
		ingredientCategories = strings.Split(ingredientCategoriesStr, ",")
	}

	ingredientsStr := r.URL.Query().Get("ingredients")
	var ingredients []string
	if ingredientsStr == "" {
		ingredients = []string{}
	} else {
		ingredients = strings.Split(ingredientsStr, ",")
	}

	dishes, err := service.NewDishService().Find(
		keyword,
		&page,
		&limit,
		&tags,
		preparationTimeFrom,
		preparationTimeTo,
		cookingTimeFrom,
		cookingTimeTo,
		&difficultLevels,
		&mealCategories,
		&ingredientCategories,
		&ingredients)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	count, err := service.NewDishService().Count(
		keyword,
		&tags,
		preparationTimeFrom,
		preparationTimeTo,
		cookingTimeFrom,
		cookingTimeTo,
		&difficultLevels,
		&mealCategories,
		&ingredientCategories,
		&ingredients)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(helper.NewPaginationHelper().PaginationJson(dishes, count))
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
