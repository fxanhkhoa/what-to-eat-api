package model

import "time"

type IngredientsInDish struct {
	Quantity     float64 `json:"quantity" bson:"quantity"`
	Slug         string  `json:"slug" bson:"slug"`
	Note         *string `json:"note,omitempty" bson:"note"`
	IngredientId string  `json:"ingredientId" bson:"ingredientId"`
}

type QueryDishDto struct {
	BaseDto
	Keyword              *string   `json:"keyword"`
	Tags                 *[]string `json:"tags"`
	PreparationTimeFrom  *int      `json:"preparationTimeFrom"`
	PreparationTimeTo    *int      `json:"preparationTimeTo"`
	CookingTimeFrom      *int      `json:"cookingTimeFrom"`
	CookingTimeTo        *int      `json:"cookingTimeTo"`
	DifficultLevels      *[]string `json:"difficultLevels"`
	MealCategories       *[]string `json:"mealCategories"`
	IngredientCategories *[]string `json:"ingredientCategories"`
	Ingredients          *[]string `json:"ingredients"`
	Labels               *[]string `json:"labels"`
}

type QueryDishRandomDto struct {
	Limit          int       `json:"limit"`
	MealCategories *[]string `json:"mealCategories"`
}

type Dish struct {
	Slug                 string               `json:"slug" bson:"slug"`
	Title                []*MultiLanguage     `json:"title" bson:"title"`
	ShortDescription     []*MultiLanguage     `json:"shortDescription" bson:"shortDescription"`
	Content              []*MultiLanguage     `json:"content" bson:"content"`
	Tags                 []*string            `json:"tags" bson:"tags"`
	PreparationTime      *float64             `json:"preparationTime,omitempty" bson:"preparationTime"`
	CookingTime          *float64             `json:"cookingTime,omitempty" bson:"cookingTime"`
	DifficultLevel       *string              `json:"difficultLevel,omitempty" bson:"difficultLevel"`
	MealCategories       []*string            `json:"mealCategories" bson:"mealCategories"`
	IngredientCategories []*string            `json:"ingredientCategories" bson:"ingredientCategories"`
	Thumbnail            *string              `json:"thumbnail,omitempty" bson:"thumbnail"`
	Videos               []*string            `json:"videos" bson:"videos"`
	Ingredients          []*IngredientsInDish `json:"ingredients" bson:"ingredients"`
	RelatedDishes        []*string            `json:"relatedDishes" bson:"relatedDishes"`
	Labels               []*string            `json:"labels" bson:"labels"`
	Deleted              bool                 `json:"deleted" bson:"deleted"`
	DeletedAt            *time.Time           `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy            *string              `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt            *time.Time           `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy            *string              `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt            *time.Time           `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy            *string              `json:"createdBy,omitempty" bson:"createdBy"`
	ID                   string               `json:"_id" bson:"_id,omitempty"`
}

type CreateDishDto struct {
	Slug                 string               `json:"slug" bson:"slug"`
	Title                []*MultiLanguage     `json:"title" bson:"title"`
	ShortDescription     []*MultiLanguage     `json:"shortDescription" bson:"shortDescription"`
	Content              []*MultiLanguage     `json:"content" bson:"content"`
	Tags                 []*string            `json:"tags" bson:"tags"`
	PreparationTime      *float64             `json:"preparationTime,omitempty" bson:"preparationTime"`
	CookingTime          *float64             `json:"cookingTime,omitempty" bson:"cookingTime"`
	DifficultLevel       *string              `json:"difficultLevel,omitempty" bson:"difficultLevel"`
	MealCategories       []*string            `json:"mealCategories" bson:"mealCategories"`
	IngredientCategories []*string            `json:"ingredientCategories" bson:"ingredientCategories"`
	Thumbnail            *string              `json:"thumbnail,omitempty" bson:"thumbnail"`
	Videos               []*string            `json:"videos" bson:"videos"`
	Ingredients          []*IngredientsInDish `json:"ingredients" bson:"ingredients"`
	RelatedDishes        []*string            `json:"relatedDishes" bson:"relatedDishes"`
	Labels               []*string            `json:"labels" bson:"labels"`
}

type UpdateDishDto struct {
	ID                   string               `json:"_id" bson:"_id,omitempty"`
	Slug                 string               `json:"slug" bson:"slug"`
	Title                []*MultiLanguage     `json:"title" bson:"title"`
	ShortDescription     []*MultiLanguage     `json:"shortDescription" bson:"shortDescription"`
	Content              []*MultiLanguage     `json:"content" bson:"content"`
	Tags                 []*string            `json:"tags" bson:"tags"`
	PreparationTime      *float64             `json:"preparationTime,omitempty" bson:"preparationTime"`
	CookingTime          *float64             `json:"cookingTime,omitempty" bson:"cookingTime"`
	DifficultLevel       *string              `json:"difficultLevel,omitempty" bson:"difficultLevel"`
	MealCategories       []*string            `json:"mealCategories" bson:"mealCategories"`
	IngredientCategories []*string            `json:"ingredientCategories" bson:"ingredientCategories"`
	Thumbnail            *string              `json:"thumbnail,omitempty" bson:"thumbnail"`
	Videos               []*string            `json:"videos" bson:"videos"`
	Ingredients          []*IngredientsInDish `json:"ingredients" bson:"ingredients"`
	RelatedDishes        []*string            `json:"relatedDishes" bson:"relatedDishes"`
	Labels               []*string            `json:"labels" bson:"labels"`
}
