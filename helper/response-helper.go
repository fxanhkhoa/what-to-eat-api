package helper

import (
	"encoding/json"
)

type ResponseHelper struct{}
type ErrorObject struct {
	ErrorStr string `json:"error"`
}

func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

func (rh *ResponseHelper) ErrorJson(errorStr string) string {
	obj := ErrorObject{
		ErrorStr: errorStr,
	}
	json, _ := json.Marshal(obj)
	return string(json)
}
