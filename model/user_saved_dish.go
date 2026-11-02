package model

import "time"

// UserSavedDish represents a user's saved recipe
type UserSavedDish struct {
	UserID    string     `json:"userId" bson:"userId"`
	DishID    string     `json:"dishId" bson:"dishId"`
	DishSlug  string     `json:"dishSlug" bson:"dishSlug"`
	Notes     *string    `json:"notes,omitempty" bson:"notes"` // User's personal notes
	Tags      []*string  `json:"tags,omitempty" bson:"tags"`   // User's custom tags
	Deleted   bool       `json:"deleted" bson:"deleted"`
	DeletedAt *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy *string    `json:"createdBy,omitempty" bson:"createdBy"`
	ID        string     `json:"_id" bson:"_id,omitempty"`
}

// QueryUserSavedDishDto is used for querying user saved dishes
type QueryUserSavedDishDto struct {
	BaseDto
	UserID   *string   `json:"userId"`
	DishSlug *string   `json:"dishSlug"`
	Tags     *[]string `json:"tags"`
}

// CreateUserSavedDishDto is used for creating a new saved dish
type CreateUserSavedDishDto struct {
	UserID   string    `json:"userId" bson:"userId"`
	DishID   string    `json:"dishId" bson:"dishId"`
	DishSlug string    `json:"dishSlug" bson:"dishSlug"`
	Notes    *string   `json:"notes,omitempty" bson:"notes"`
	Tags     []*string `json:"tags,omitempty" bson:"tags"`
}

// UpdateUserSavedDishDto is used for updating a saved dish
type UpdateUserSavedDishDto struct {
	ID       string    `json:"_id" bson:"_id,omitempty"`
	UserID   string    `json:"userId" bson:"userId"`
	DishSlug string    `json:"dishSlug" bson:"dishSlug"`
	Notes    *string   `json:"notes,omitempty" bson:"notes"`
	Tags     []*string `json:"tags,omitempty" bson:"tags"`
}

// DeleteUserSavedDishDto is used for removing a saved dish
type DeleteUserSavedDishDto struct {
	UserID   string `json:"userId" bson:"userId"`
	DishSlug string `json:"dishSlug" bson:"dishSlug"`
}
