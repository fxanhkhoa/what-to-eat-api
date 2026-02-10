package model

import "time"

// UserDishInteraction tracks user interactions with dishes for personalization
type UserDishInteraction struct {
	UserID          string     `json:"userId" bson:"userId"`
	DishID          string     `json:"dishId" bson:"dishId"`
	DishSlug        string     `json:"dishSlug" bson:"dishSlug"`
	ViewCount       int        `json:"viewCount" bson:"viewCount"`             // Number of times user viewed this dish
	LastViewedAt    *time.Time `json:"lastViewedAt,omitempty" bson:"lastViewedAt"` // Last time user viewed
	Cooked          bool       `json:"cooked" bson:"cooked"`                   // Has user marked as cooked
	CookedCount     int        `json:"cookedCount" bson:"cookedCount"`         // Number of times marked as cooked
	LastCookedAt    *time.Time `json:"lastCookedAt,omitempty" bson:"lastCookedAt"`
	Rating          *int       `json:"rating,omitempty" bson:"rating"`         // User's rating (1-5)
	RatedAt         *time.Time `json:"ratedAt,omitempty" bson:"ratedAt"`
	SharedCount     int        `json:"sharedCount" bson:"sharedCount"`         // Number of times shared
	LastSharedAt    *time.Time `json:"lastSharedAt,omitempty" bson:"lastSharedAt"`
	InteractionScore float64   `json:"interactionScore" bson:"interactionScore"` // Calculated score based on all interactions
	Deleted         bool       `json:"deleted" bson:"deleted"`
	DeletedAt       *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy       *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt       *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy       *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt       *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy       *string    `json:"createdBy,omitempty" bson:"createdBy"`
	ID              string     `json:"_id" bson:"_id,omitempty"`
}

// QueryUserDishInteractionDto is used for querying user interactions
type QueryUserDishInteractionDto struct {
	BaseDto
	UserID     *string `json:"userId"`
	DishSlug   *string `json:"dishSlug"`
	Cooked     *bool   `json:"cooked"`
	MinRating  *int    `json:"minRating"`
}

// RecordDishViewDto is used for recording a dish view
type RecordDishViewDto struct {
	UserID   string `json:"userId" bson:"userId"`
	DishID   string `json:"dishId" bson:"dishId"`
	DishSlug string `json:"dishSlug" bson:"dishSlug"`
}

// RecordDishCookedDto is used for marking a dish as cooked
type RecordDishCookedDto struct {
	UserID   string `json:"userId" bson:"userId"`
	DishID   string `json:"dishId" bson:"dishId"`
	DishSlug string `json:"dishSlug" bson:"dishSlug"`
}

// RateDishDto is used for rating a dish
type RateDishDto struct {
	UserID   string `json:"userId" bson:"userId"`
	DishID   string `json:"dishId" bson:"dishId"`
	DishSlug string `json:"dishSlug" bson:"dishSlug"`
	Rating   int    `json:"rating" bson:"rating"` // 1-5
}

// RecordDishSharedDto is used for tracking when a dish is shared
type RecordDishSharedDto struct {
	UserID   string `json:"userId" bson:"userId"`
	DishID   string `json:"dishId" bson:"dishId"`
	DishSlug string `json:"dishSlug" bson:"dishSlug"`
}
