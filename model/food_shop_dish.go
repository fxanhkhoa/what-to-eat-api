package model

import "time"

// FoodShopDish represents the relationship between a food shop and a dish it serves
type FoodShopDish struct {
	ID          string     `json:"_id" bson:"_id,omitempty"`
	FoodShopID  string     `json:"foodShopId" bson:"foodShopId"`
	DishSlug    string     `json:"dishSlug" bson:"dishSlug"`
	Price       *float64   `json:"price,omitempty" bson:"price"`
	Currency    *string    `json:"currency,omitempty" bson:"currency"` // e.g. "VND", "USD"
	IsAvailable bool       `json:"isAvailable" bson:"isAvailable"`
	Deleted     bool       `json:"deleted" bson:"deleted"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy   *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy   *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt   *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy   *string    `json:"createdBy,omitempty" bson:"createdBy"`
}

// CreateFoodShopDishDto is used for linking a dish to a food shop
type CreateFoodShopDishDto struct {
	FoodShopID  string   `json:"foodShopId" bson:"foodShopId"`
	DishSlug    string   `json:"dishSlug" bson:"dishSlug"`
	Price       *float64 `json:"price,omitempty" bson:"price"`
	Currency    *string  `json:"currency,omitempty" bson:"currency"`
	IsAvailable bool     `json:"isAvailable" bson:"isAvailable"`
}

// UpdateFoodShopDishDto is used for updating a food shop dish entry
type UpdateFoodShopDishDto struct {
	ID          string   `json:"_id" bson:"_id,omitempty"`
	Price       *float64 `json:"price,omitempty" bson:"price"`
	Currency    *string  `json:"currency,omitempty" bson:"currency"`
	IsAvailable *bool    `json:"isAvailable,omitempty" bson:"isAvailable"`
}

// QueryFoodShopDishDto is used for querying food shop dish entries
type QueryFoodShopDishDto struct {
	BaseDto
	FoodShopID *string `json:"foodShopId"`
	DishSlug   *string `json:"dishSlug"`
}
