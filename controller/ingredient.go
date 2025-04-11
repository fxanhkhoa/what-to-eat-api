package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"what-to-eat/be/graph/service"
	"what-to-eat/be/helper"

	"github.com/gorilla/mux"
)

type IngredientController struct{}

func NewIngredientController() *IngredientController {
	return &IngredientController{}
}

func (ic *IngredientController) Find(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, errPage := strconv.Atoi(pageStr)
	if errPage != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(errPage.Error()), http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}

	limit, errLimit := strconv.Atoi(limitStr)
	if errLimit != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(errLimit.Error()), http.StatusBadRequest)
		return
	}

	keywordStr := r.URL.Query().Get("keyword")
	var keyword *string
	if keywordStr == "" {
		keyword = nil
	} else {
		keyword = &keywordStr
	}

	ingredients, err := service.NewIngredientService().Find(keyword, &page, &limit)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	count, err := service.NewIngredientService().Count(keyword)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(helper.NewPaginationHelper().PaginationJson(ingredients, count))
}

func (ic *IngredientController) FindOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ingredient, err := service.NewIngredientService().FindOne(vars["slug"])
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	decoded, err := json.Marshal(ingredient)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(decoded)
}
