package helper

import (
	"encoding/json"
)

type PaginationHelper struct{}

type PaginationObject struct {
	Data  any   `json:"data"`
	Count int64 `json:"count"`
}

func NewPaginationHelper() *PaginationHelper {
	return &PaginationHelper{}
}

func (ph *PaginationHelper) PaginationJson(data any, count int64) []byte {
	obj := PaginationObject{
		Data:  data,
		Count: count,
	}
	json, _ := json.Marshal(obj)
	return json
}
