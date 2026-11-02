package model

import "time"

// UserDishLocation represents a personal (private) location record added by a user for a dish
type UserDishLocation struct {
	ID         string     `json:"_id" bson:"_id,omitempty"`
	UserID     string     `json:"userId" bson:"userId"`
	DishSlug   string     `json:"dishSlug" bson:"dishSlug"`
	FoodShopID *string    `json:"foodShopId,omitempty" bson:"foodShopId"` // optional ref to an existing FoodShop
	Name       string     `json:"name" bson:"name"`                        // custom place name
	Address    *string    `json:"address,omitempty" bson:"address"`
	Latitude   *float64   `json:"latitude,omitempty" bson:"latitude"`
	Longitude  *float64   `json:"longitude,omitempty" bson:"longitude"`
	Notes      *string    `json:"notes,omitempty" bson:"notes"` // personal note
	Deleted    bool       `json:"deleted" bson:"deleted"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy  *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy  *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt  *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy  *string    `json:"createdBy,omitempty" bson:"createdBy"`
}

// CreateUserDishLocationDto is used for adding a personal location for a dish
type CreateUserDishLocationDto struct {
	UserID     string   `json:"userId" bson:"userId"`
	DishSlug   string   `json:"dishSlug" bson:"dishSlug"`
	FoodShopID *string  `json:"foodShopId,omitempty" bson:"foodShopId"`
	Name       string   `json:"name" bson:"name"`
	Address    *string  `json:"address,omitempty" bson:"address"`
	Latitude   *float64 `json:"latitude,omitempty" bson:"latitude"`
	Longitude  *float64 `json:"longitude,omitempty" bson:"longitude"`
	Notes      *string  `json:"notes,omitempty" bson:"notes"`
}

// UpdateUserDishLocationDto is used for updating a personal dish location
type UpdateUserDishLocationDto struct {
	ID        string   `json:"_id" bson:"_id,omitempty"`
	Name      *string  `json:"name,omitempty" bson:"name"`
	Address   *string  `json:"address,omitempty" bson:"address"`
	Latitude  *float64 `json:"latitude,omitempty" bson:"latitude"`
	Longitude *float64 `json:"longitude,omitempty" bson:"longitude"`
	Notes     *string  `json:"notes,omitempty" bson:"notes"`
}

// QueryUserDishLocationDto is used for querying personal dish locations
type QueryUserDishLocationDto struct {
	BaseDto
	UserID   *string `json:"userId"`
	DishSlug *string `json:"dishSlug"`
}
