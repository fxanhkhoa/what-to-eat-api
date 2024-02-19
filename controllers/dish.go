package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"what-to-eat/be/graph/service"
	"what-to-eat/be/helper"

	"github.com/gorilla/mux"
)

type DishController struct{}

func NewDishController() *DishController {
	return &DishController{}
}

func (dc *DishController) Find(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, errPage := strconv.Atoi(pageStr)
	if errPage != nil {
		http.Error(w, errPage.Error(), http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}

	limit, errLimit := strconv.Atoi(limitStr)
	if errLimit != nil {
		http.Error(w, errLimit.Error(), http.StatusBadRequest)
		return
	}

	keywordStr := r.URL.Query().Get("keyword")
	var keyword *string
	if keywordStr == "" {
		keyword = nil
	} else {
		keyword = &keywordStr
	}

	dishes, err := service.NewDishService().Find(keyword, &page, &limit)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	count, err := service.NewDishService().Count(keyword)
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
