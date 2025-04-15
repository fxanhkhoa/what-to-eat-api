package model

import "time"

type Ingredient struct {
	Slug               string           `json:"slug" bson:"slug"`
	Title              []*MultiLanguage `json:"title" bson:"title"`
	Measure            *string          `json:"measure,omitempty" bson:"measure"`
	Calories           *int             `json:"calories,omitempty" bson:"calories"`
	Carbohydrate       *int             `json:"carbohydrate,omitempty" bson:"carbohydrate"`
	Fat                *int             `json:"fat,omitempty" bson:"fat"`
	IngredientCategory []*string        `json:"ingredientCategory" bson:"ingredientCategory"`
	Weight             *int             `json:"weight,omitempty" bson:"weight"`
	Protein            *int             `json:"protein,omitempty" bson:"protein"`
	Cholesterol        *int             `json:"cholesterol,omitempty" bson:"cholesterol"`
	Sodium             *int             `json:"sodium,omitempty" bson:"sodium"`
	Images             []*string        `json:"images" bson:"images"`
	Deleted            bool             `json:"deleted" bson:"deleted"`
	DeletedAt          *time.Time       `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy          *string          `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt          *time.Time       `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy          *string          `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt          *time.Time       `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy          *string          `json:"createdBy,omitempty" bson:"createdBy"`
	ID                 string           `json:"_id" bson:"_id,omitempty"`
}

type QueryIngredientDto struct {
	BaseDto
	Keyword *string `json:"keyword"`
}

type CreateIngredientDto struct {
	Slug               string           `json:"slug" bson:"slug"`
	Title              []*MultiLanguage `json:"title" bson:"title"`
	Measure            *string          `json:"measure,omitempty" bson:"measure"`
	Calories           *int             `json:"calories,omitempty" bson:"calories"`
	Carbohydrate       *int             `json:"carbohydrate,omitempty" bson:"carbohydrate"`
	Fat                *int             `json:"fat,omitempty" bson:"fat"`
	IngredientCategory []*string        `json:"ingredientCategory" bson:"ingredientCategory"`
	Weight             *int             `json:"weight,omitempty" bson:"weight"`
	Protein            *int             `json:"protein,omitempty" bson:"protein"`
	Cholesterol        *int             `json:"cholesterol,omitempty" bson:"cholesterol"`
	Sodium             *int             `json:"sodium,omitempty" bson:"sodium"`
	Images             []*string        `json:"images" bson:"images"`
}

type UpdateIngredientDto struct {
	ID                 string           `json:"_id" bson:"_id,omitempty"`
	Slug               string           `json:"slug" bson:"slug"`
	Title              []*MultiLanguage `json:"title" bson:"title"`
	Measure            *string          `json:"measure,omitempty" bson:"measure"`
	Calories           *int             `json:"calories,omitempty" bson:"calories"`
	Carbohydrate       *int             `json:"carbohydrate,omitempty" bson:"carbohydrate"`
	Fat                *int             `json:"fat,omitempty" bson:"fat"`
	IngredientCategory []*string        `json:"ingredientCategory" bson:"ingredientCategory"`
	Images             []*string        `json:"images" bson:"images"`
	Weight             *int             `json:"weight,omitempty" bson:"weight"`
	Protein            *int             `json:"protein,omitempty" bson:"protein"`
	Cholesterol        *int             `json:"cholesterol,omitempty" bson:"cholesterol"`
	Sodium             *int             `json:"sodium,omitempty" bson:"sodium"`
}
