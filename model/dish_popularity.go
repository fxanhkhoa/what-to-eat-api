package model

import "time"

// DishPopularity tracks dish popularity metrics for trending recommendations
type DishPopularity struct {
	DishID              string     `json:"dishId" bson:"dishId"`
	DishSlug            string     `json:"dishSlug" bson:"dishSlug"`
	TotalViews          int64      `json:"totalViews" bson:"totalViews"`                   // Total number of views
	ViewsLast7Days      int64      `json:"viewsLast7Days" bson:"viewsLast7Days"`           // Views in last 7 days
	ViewsLast30Days     int64      `json:"viewsLast30Days" bson:"viewsLast30Days"`         // Views in last 30 days
	TotalFavorites      int64      `json:"totalFavorites" bson:"totalFavorites"`           // Total number of favorites
	TotalSaves          int64      `json:"totalSaves" bson:"totalSaves"`                   // Total number of saves
	TotalCookedCount    int64      `json:"totalCookedCount" bson:"totalCookedCount"`       // Total times marked as cooked
	TotalShares         int64      `json:"totalShares" bson:"totalShares"`                 // Total times shared
	AverageRating       float64    `json:"averageRating" bson:"averageRating"`             // Average rating
	TotalRatings        int64      `json:"totalRatings" bson:"totalRatings"`               // Number of ratings
	TrendingScore       float64    `json:"trendingScore" bson:"trendingScore"`             // Calculated trending score
	PopularityScore     float64    `json:"popularityScore" bson:"popularityScore"`         // Overall popularity score
	LastCalculatedAt    *time.Time `json:"lastCalculatedAt,omitempty" bson:"lastCalculatedAt"` // Last time scores were calculated
	Deleted             bool       `json:"deleted" bson:"deleted"`
	DeletedAt           *time.Time `json:"deletedAt,omitempty" bson:"deletedAt"`
	DeletedBy           *string    `json:"deletedBy,omitempty" bson:"deletedBy"`
	UpdatedAt           *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	UpdatedBy           *string    `json:"updatedBy,omitempty" bson:"updatedBy"`
	CreatedAt           *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	CreatedBy           *string    `json:"createdBy,omitempty" bson:"createdBy"`
	ID                  string     `json:"_id" bson:"_id,omitempty"`
}

// QueryDishPopularityDto is used for querying popular dishes
type QueryDishPopularityDto struct {
	BaseDto
	MinAverageRating    *float64 `json:"minAverageRating"`
	MinTrendingScore    *float64 `json:"minTrendingScore"`
	MinPopularityScore  *float64 `json:"minPopularityScore"`
	SortBy              *string  `json:"sortBy"` // "trending", "popular", "rating", "views"
}

// TrendingDishesDto is used for getting trending dishes
type TrendingDishesDto struct {
	Limit          int       `json:"limit"`
	MealCategories *[]string `json:"mealCategories"`
	Period         string    `json:"period"` // "day", "week", "month"
}
