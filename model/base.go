package model

type BaseDto struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type LanguageStringData struct {
	Lang string `json:"lang" bson:"lang,omitempty"`
	Data string `json:"data" bson:"data,omitempty"`
}

type CountMetaData struct {
	TotalItems   int64   `json:"totalItems"`
	ItemCount    int     `json:"itemCount"`
	ItemsPerPage int     `json:"itemsPerPage"`
	TotalPages   float64 `json:"totalPages"`
	CurrentPage  int     `json:"currentPage"`
}

type PaginationResponse struct {
	Data     any           `json:"data"`
	Metadata CountMetaData `json:"metadata"`
}

type MultiLanguage struct {
	Lang string  `json:"lang" bson:"lang"`
	Data *string `json:"data,omitempty" bson:"data"`
}
