package model

import "time"

// UserDietaryPreference represents a user's dietary preferences and restrictions
type UserDietaryPreference struct {
	UserID              string     `json:"userId" bson:"userId"`
	DietaryRestrictions []*string  `json:"dietaryRestrictions,omitempty" bson:"dietaryRestrictions"` // e.g., "vegetarian", "vegan", "gluten-free", "dairy-free", "nut-free"
	Allergies           []*string  `json:"allergies,omitempty" bson:"allergies"`                     // e.g., "peanuts", "shellfish", "eggs"
	DislikedIngredients []*string  `json:"dislikedIngredients,omitempty" bson:"dislikedIngredients"` // Ingredient slugs user doesn't like
	PreferredCuisines   []*string  `json:"preferredCuisines,omitempty" bson:"preferredCuisines"`     // e.g., "italian", "japanese", "mexican"
	PreferredLabels     []*string  `json:"preferredLabels,omitempty" bson:"preferredLabels"`         // e.g., "low-calorie", "high-protein", "quick"
	SpiceLevel          *string    `json:"spiceLevel,omitempty" bson:"spiceLevel"`                   // e.g., "mild", "medium", "hot"
	DifficultLevel      []*string  `json:"difficultLevel,omitempty" bson:"difficultLevel"`           // e.g., "easy", "medium", "hard"
	MaxCookingTime      *int       `json:"maxCookingTime,omitempty" bson:"maxCookingTime"`           // in minutes
	MaxPreparationTime  *int       `json:"maxPreparationTime,omitempty" bson:"maxPreparationTime"`   // in minutes
	Deleted             bool       `json:"deleted" bson:"deleted"`
	DeletedAt           *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy           *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt           *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy           *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt           *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy           *string    `json:"createdBy,omitempty" bson:"createdBy"`
	ID                  string     `json:"_id" bson:"_id,omitempty"`
}

// CreateUserDietaryPreferenceDto is used for creating dietary preferences
type CreateUserDietaryPreferenceDto struct {
	UserID              string    `json:"userId" bson:"userId"`
	DietaryRestrictions []*string `json:"dietaryRestrictions,omitempty" bson:"dietaryRestrictions"`
	Allergies           []*string `json:"allergies,omitempty" bson:"allergies"`
	DislikedIngredients []*string `json:"dislikedIngredients,omitempty" bson:"dislikedIngredients"`
	PreferredCuisines   []*string `json:"preferredCuisines,omitempty" bson:"preferredCuisines"`
	PreferredLabels     []*string `json:"preferredLabels,omitempty" bson:"preferredLabels"`
	SpiceLevel          *string   `json:"spiceLevel,omitempty" bson:"spiceLevel"`
	DifficultLevel      []*string `json:"difficultLevel,omitempty" bson:"difficultLevel"`
	MaxCookingTime      *int      `json:"maxCookingTime,omitempty" bson:"maxCookingTime"`
	MaxPreparationTime  *int      `json:"maxPreparationTime,omitempty" bson:"maxPreparationTime"`
}

// UpdateUserDietaryPreferenceDto is used for updating dietary preferences
type UpdateUserDietaryPreferenceDto struct {
	ID                  string    `json:"_id" bson:"_id,omitempty"`
	UserID              string    `json:"userId" bson:"userId"`
	DietaryRestrictions []*string `json:"dietaryRestrictions,omitempty" bson:"dietaryRestrictions"`
	Allergies           []*string `json:"allergies,omitempty" bson:"allergies"`
	DislikedIngredients []*string `json:"dislikedIngredients,omitempty" bson:"dislikedIngredients"`
	PreferredCuisines   []*string `json:"preferredCuisines,omitempty" bson:"preferredCuisines"`
	PreferredLabels     []*string `json:"preferredLabels,omitempty" bson:"preferredLabels"`
	SpiceLevel          *string   `json:"spiceLevel,omitempty" bson:"spiceLevel"`
	DifficultLevel      []*string `json:"difficultLevel,omitempty" bson:"difficultLevel"`
	MaxCookingTime      *int      `json:"maxCookingTime,omitempty" bson:"maxCookingTime"`
	MaxPreparationTime  *int      `json:"maxPreparationTime,omitempty" bson:"maxPreparationTime"`
}
