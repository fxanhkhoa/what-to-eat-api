package model

// PersonalizedRandomDishDto is used for getting personalized random dish recommendations
type PersonalizedRandomDishDto struct {
	UserID                  string    `json:"userId" bson:"userId"`
	Limit                   int       `json:"limit"`
	MealCategories          *[]string `json:"mealCategories"`
	ExcludeDishSlugs        *[]string `json:"excludeDishSlugs"`        // Exclude specific dishes
	IncludeFavorites        bool      `json:"includeFavorites"`        // Include user favorites in pool
	IncludeSaved            bool      `json:"includeSaved"`            // Include user saved dishes in pool
	IncludeTrending         bool      `json:"includeTrending"`         // Include trending dishes
	ApplyDietaryPreferences bool      `json:"applyDietaryPreferences"` // Apply user's dietary preferences
	FavoriteWeight          float64   `json:"favoriteWeight"`          // Weight for favorite dishes (0-1)
	SavedWeight             float64   `json:"savedWeight"`             // Weight for saved dishes (0-1)
	InteractionWeight       float64   `json:"interactionWeight"`       // Weight for highly interacted dishes (0-1)
	TrendingWeight          float64   `json:"trendingWeight"`          // Weight for trending dishes (0-1)
	RatingWeight            float64   `json:"ratingWeight"`            // Weight for highly rated dishes (0-1)
	RandomnessLevel         float64   `json:"randomnessLevel"`         // 0 = fully personalized, 1 = fully random
}

// PersonalizedRecommendationResult contains the recommended dishes with scores
type PersonalizedRecommendationResult struct {
	DishSlug            string  `json:"dishSlug"`
	PersonalizationScore float64 `json:"personalizationScore"` // How well it matches user preferences
	ReasonTags          []string `json:"reasonTags"`           // Why this dish was recommended (e.g., "favorite", "trending", "matches-preference")
}

// UserPreferenceSummary contains aggregated user preference data
type UserPreferenceSummary struct {
	UserID                    string    `json:"userId"`
	TotalFavorites            int       `json:"totalFavorites"`
	TotalSaved                int       `json:"totalSaved"`
	TopInteractedDishSlugs    []string  `json:"topInteractedDishSlugs"`    // Top dishes by interaction
	PreferredMealCategories   []string  `json:"preferredMealCategories"`   // From interaction history
	PreferredIngredientCats   []string  `json:"preferredIngredientCats"`   // From interaction history
	PreferredDifficultLevels  []string  `json:"preferredDifficultLevels"`  // From dietary preferences
	DietaryRestrictions       []string  `json:"dietaryRestrictions"`       // From dietary preferences
	DislikedIngredientSlugs   []string  `json:"dislikedIngredientSlugs"`   // From dietary preferences
	AverageRatingGiven        float64   `json:"averageRatingGiven"`        // Average rating user gives
	HasDietaryPreferences     bool      `json:"hasDietaryPreferences"`
}
