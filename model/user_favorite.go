package model

import "time"

// UserFavorite represents a user's favorite dish
type UserFavorite struct {
	UserID    string     `json:"userId" bson:"userId"`
	DishID    string     `json:"dishId" bson:"dishId"`
	DishSlug  string     `json:"dishSlug" bson:"dishSlug"`
	Deleted   bool       `json:"deleted" bson:"deleted"`
	DeletedAt *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy *string    `json:"createdBy,omitempty" bson:"createdBy"`
	ID        string     `json:"_id" bson:"_id,omitempty"`
}

// QueryUserFavoriteDto is used for querying user favorites
type QueryUserFavoriteDto struct {
	BaseDto
	UserID   *string `json:"userId"`
	DishSlug *string `json:"dishSlug"`
}

// CreateUserFavoriteDto is used for creating a new favorite
type CreateUserFavoriteDto struct {
	UserID   string `json:"userId" bson:"userId"`
	DishID   string `json:"dishId" bson:"dishId"`
	DishSlug string `json:"dishSlug" bson:"dishSlug"`
}

// DeleteUserFavoriteDto is used for removing a favorite
type DeleteUserFavoriteDto struct {
	UserID   string `json:"userId" bson:"userId"`
	DishSlug string `json:"dishSlug" bson:"dishSlug"`
}
