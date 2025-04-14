package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
	"what-to-eat/be/auth"
	"what-to-eat/be/graph/model"
	"what-to-eat/be/graph/service"
	"what-to-eat/be/helper"

	"github.com/gorilla/mux"
)

type DishVoteController struct{}

func NewDishVoteController() *DishVoteController {
	return &DishVoteController{}
}

func (dc *DishVoteController) Find(w http.ResponseWriter, r *http.Request) {
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

	dishes, err := service.NewDishVoteService().Find(
		keyword,
		&page,
		&limit)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	count, err := service.NewDishVoteService().Count(keyword)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(helper.NewPaginationHelper().PaginationJson(dishes, count))
}

func (dvc *DishVoteController) FindOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dishVote, err := service.NewDishVoteService().FindOne(vars["id"])
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	decoded, err := json.Marshal(dishVote)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(decoded)
}

func (dvc *DishVoteController) Create(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	var t model.CreateDishVoteInput
	err := json.Unmarshal(data, &t)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := auth.ForContext(r.Context())
	if user == nil {
		user = &model.User{
			Email:       "",
			Password:    new(string),
			Name:        new(string),
			DateOfBirth: &time.Time{},
			Address:     new(string),
			Phone:       new(string),
			GoogleID:    new(string),
			FacebookID:  new(string),
			GithubID:    new(string),
			Avatar:      new(string),
			Deleted:     false,
			DeletedAt:   &time.Time{},
			DeletedBy:   new(string),
			UpdatedAt:   &time.Time{},
			UpdatedBy:   new(string),
			CreatedAt:   &time.Time{},
			CreatedBy:   new(string),
			ID:          "",
			RoleName:    "",
		}
	}
	result, err := service.NewDishVoteService().Create(t, user)

	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}

	decoded, err := json.Marshal(result)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(decoded)
}

func (dvc *DishVoteController) Update(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	var t model.UpdateDishVoteInput
	err := json.Unmarshal(data, &t)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	t.ID = vars["id"]
	user := auth.ForContext(r.Context())
	if user == nil {
		user = &model.User{
			Email:       "",
			Password:    new(string),
			Name:        new(string),
			DateOfBirth: &time.Time{},
			Address:     new(string),
			Phone:       new(string),
			GoogleID:    new(string),
			FacebookID:  new(string),
			GithubID:    new(string),
			Avatar:      new(string),
			Deleted:     false,
			DeletedAt:   &time.Time{},
			DeletedBy:   new(string),
			UpdatedAt:   &time.Time{},
			UpdatedBy:   new(string),
			CreatedAt:   &time.Time{},
			CreatedBy:   new(string),
			ID:          "",
			RoleName:    "",
		}
	}
	result, err := service.NewDishVoteService().Update(t, user)

	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}

	decoded, err := json.Marshal(result)
	if err != nil {
		http.Error(w, helper.NewResponseHelper().ErrorJson(err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(decoded)
}
