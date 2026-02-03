package service

import (
	"context"
	"math"
	"math/rand"
	"time"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PersonalizedRandomService struct {
	favoriteService          *UserFavoriteService
	savedDishService         *UserSavedDishService
	dietaryPreferenceService *UserDietaryPreferenceService
	interactionService       *UserDishInteractionService
	popularityService        *DishPopularityService
	collectionService        *UserDishCollectionService
	dishService              *DishService
}

func NewPersonalizedRandomService() *PersonalizedRandomService {
	return &PersonalizedRandomService{
		favoriteService:          NewUserFavoriteService(),
		savedDishService:         NewUserSavedDishService(),
		dietaryPreferenceService: NewUserDietaryPreferenceService(),
		interactionService:       NewUserDishInteractionService(),
		popularityService:        NewDishPopularityService(),
		collectionService:        NewUserDishCollectionService(),
		dishService:              &DishService{},
	}
}

func (s *PersonalizedRandomService) GetPersonalizedRandomDishes(dto model.PersonalizedRandomDishDto) ([]model.PersonalizedRecommendationResult, error) {
	// Get user preference summary
	preferenceSummary, err := s.GetUserPreferenceSummary(dto.UserID)
	if err != nil {
		return nil, err
	}

	// Build dish filter based on preferences
	filter := bson.M{"deleted": false}
	
	// Apply meal categories filter
	if dto.MealCategories != nil && len(*dto.MealCategories) > 0 {
		filter["mealCategories"] = bson.M{"$in": *dto.MealCategories}
	}

	// Exclude specific dishes
	if dto.ExcludeDishSlugs != nil && len(*dto.ExcludeDishSlugs) > 0 {
		filter["slug"] = bson.M{"$nin": *dto.ExcludeDishSlugs}
	}

	// Apply dietary preferences
	if dto.ApplyDietaryPreferences && preferenceSummary.HasDietaryPreferences {
		dietaryPref, _ := s.dietaryPreferenceService.Get(dto.UserID)
		if dietaryPref != nil {
			// Filter by dietary restrictions
			if dietaryPref.DietaryRestrictions != nil && len(dietaryPref.DietaryRestrictions) > 0 {
				filter["labels"] = bson.M{"$in": dietaryPref.DietaryRestrictions}
			}

			// Exclude disliked ingredients
			if dietaryPref.DislikedIngredients != nil && len(dietaryPref.DislikedIngredients) > 0 {
				filter["ingredients.slug"] = bson.M{"$nin": dietaryPref.DislikedIngredients}
			}

			// Filter by preferred labels
			if dietaryPref.PreferredLabels != nil && len(dietaryPref.PreferredLabels) > 0 {
				filter["$or"] = []bson.M{
					{"labels": bson.M{"$in": dietaryPref.PreferredLabels}},
					{"labels": bson.M{"$exists": false}},
				}
			}

			// Filter by difficulty level
			if dietaryPref.DifficultLevel != nil && len(dietaryPref.DifficultLevel) > 0 {
				filter["difficultLevel"] = bson.M{"$in": dietaryPref.DifficultLevel}
			}

			// Filter by max cooking time
			if dietaryPref.MaxCookingTime != nil {
				filter["cookingTime"] = bson.M{"$lte": *dietaryPref.MaxCookingTime}
			}

			// Filter by max preparation time
			if dietaryPref.MaxPreparationTime != nil {
				filter["preparationTime"] = bson.M{"$lte": *dietaryPref.MaxPreparationTime}
			}
		}
	}

	// Get candidate dishes
	opts := options.Find().SetLimit(int64(dto.Limit * 10)) // Get more than needed for scoring
	cursor, err := s.dishService.Collection().Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var dishes []model.Dish
	if err = cursor.All(context.TODO(), &dishes); err != nil {
		return nil, err
	}

	// Score each dish
	scoredDishes := make([]struct {
		dish   model.Dish
		score  float64
		reason []string
	}, 0, len(dishes))

	for _, dish := range dishes {
		score, reasons := s.calculatePersonalizationScore(dish, preferenceSummary, dto)
		scoredDishes = append(scoredDishes, struct {
			dish   model.Dish
			score  float64
			reason []string
		}{dish, score, reasons})
	}

	// Apply randomness level
	if dto.RandomnessLevel > 0 {
		// Add random component to scores
		rand.Seed(time.Now().UnixNano())
		for i := range scoredDishes {
			randomFactor := rand.Float64() * dto.RandomnessLevel
			scoredDishes[i].score = scoredDishes[i].score*(1-dto.RandomnessLevel) + randomFactor*100
		}
	}

	// Sort by score (descending)
	for i := 0; i < len(scoredDishes); i++ {
		for j := i + 1; j < len(scoredDishes); j++ {
			if scoredDishes[j].score > scoredDishes[i].score {
				scoredDishes[i], scoredDishes[j] = scoredDishes[j], scoredDishes[i]
			}
		}
	}

	// Take top N dishes
	limit := dto.Limit
	if limit > len(scoredDishes) {
		limit = len(scoredDishes)
	}

	results := make([]model.PersonalizedRecommendationResult, limit)
	for i := 0; i < limit; i++ {
		results[i] = model.PersonalizedRecommendationResult{
			DishSlug:             scoredDishes[i].dish.Slug,
			PersonalizationScore: scoredDishes[i].score,
			ReasonTags:           scoredDishes[i].reason,
		}
	}

	return results, nil
}

func (s *PersonalizedRandomService) calculatePersonalizationScore(
	dish model.Dish,
	preferenceSummary *model.UserPreferenceSummary,
	dto model.PersonalizedRandomDishDto,
) (float64, []string) {
	score := 0.0
	reasons := []string{}

	// Check if dish is in favorites
	if dto.IncludeFavorites {
		for _, slug := range preferenceSummary.TopInteractedDishSlugs {
			if slug == dish.Slug {
				score += 100 * dto.FavoriteWeight
				reasons = append(reasons, "favorite")
				break
			}
		}
	}

	// Get interaction data for this dish
	interactionFilter := bson.M{"userId": dto.UserID, "dishSlug": dish.Slug, "deleted": false}
	var interaction model.UserDishInteraction
	s.interactionService.Collection().FindOne(context.TODO(), interactionFilter).Decode(&interaction)
	
	if interaction.ID != "" {
		// Has interacted with this dish
		score += interaction.InteractionScore * dto.InteractionWeight
		if interaction.ViewCount > 0 {
			reasons = append(reasons, "previously-viewed")
		}
		if interaction.Cooked {
			reasons = append(reasons, "previously-cooked")
		}
		if interaction.Rating != nil && *interaction.Rating >= 4 {
			reasons = append(reasons, "highly-rated-by-you")
		}
	}

	// Check popularity/trending
	if dto.IncludeTrending {
		popularity, _ := s.popularityService.GetPopularityMetrics(dish.Slug)
		if popularity != nil {
			score += popularity.TrendingScore * dto.TrendingWeight
			score += popularity.AverageRating * 10 * dto.RatingWeight
			
			if popularity.TrendingScore > 50 {
				reasons = append(reasons, "trending")
			}
			if popularity.AverageRating >= 4.0 {
				reasons = append(reasons, "highly-rated")
			}
		}
	}

	// Match with preferred meal categories
	if dish.MealCategories != nil && len(preferenceSummary.PreferredMealCategories) > 0 {
		for _, mealCat := range dish.MealCategories {
			for _, prefCat := range preferenceSummary.PreferredMealCategories {
				if mealCat != nil && *mealCat == prefCat {
					score += 20
					reasons = append(reasons, "matches-preference")
					break
				}
			}
		}
	}

	// Match with preferred ingredient categories
	if dish.IngredientCategories != nil && len(preferenceSummary.PreferredIngredientCats) > 0 {
		for _, ingCat := range dish.IngredientCategories {
			for _, prefCat := range preferenceSummary.PreferredIngredientCats {
				if ingCat != nil && *ingCat == prefCat {
					score += 15
					break
				}
			}
		}
	}

	// Bonus for quick dishes if that's a preference
	if dish.PreparationTime != nil && dish.CookingTime != nil {
		totalTime := *dish.PreparationTime + *dish.CookingTime
		if totalTime <= 30 {
			score += 10
			reasons = append(reasons, "quick-to-make")
		}
	}

	// Normalize score to 0-100 range
	score = math.Min(score, 100)

	return score, reasons
}

func (s *PersonalizedRandomService) GetUserPreferenceSummary(userID string) (*model.UserPreferenceSummary, error) {
	summary := &model.UserPreferenceSummary{
		UserID: userID,
	}

	// Get favorite count
	favoriteCount, _ := s.favoriteService.GetFavoriteCount(userID)
	summary.TotalFavorites = int(favoriteCount)

	// Get saved dish count
	savedCount, _ := s.savedDishService.GetSavedCount(userID)
	summary.TotalSaved = int(savedCount)

	// Get top interacted dishes
	topDishes, _ := s.interactionService.GetTopInteractedDishes(userID, 20)
	summary.TopInteractedDishSlugs = topDishes

	// Get average rating given by user
	avgRating, _ := s.interactionService.GetAverageRating(userID)
	summary.AverageRatingGiven = avgRating

	// Get dietary preferences
	dietaryPref, _ := s.dietaryPreferenceService.Get(userID)
	if dietaryPref != nil {
		summary.HasDietaryPreferences = true
		
		if dietaryPref.DietaryRestrictions != nil {
			for _, restriction := range dietaryPref.DietaryRestrictions {
				if restriction != nil {
					summary.DietaryRestrictions = append(summary.DietaryRestrictions, *restriction)
				}
			}
		}
		
		if dietaryPref.DislikedIngredients != nil {
			for _, ing := range dietaryPref.DislikedIngredients {
				if ing != nil {
					summary.DislikedIngredientSlugs = append(summary.DislikedIngredientSlugs, *ing)
				}
			}
		}

		if dietaryPref.DifficultLevel != nil {
			for _, level := range dietaryPref.DifficultLevel {
				if level != nil {
					summary.PreferredDifficultLevels = append(summary.PreferredDifficultLevels, *level)
				}
			}
		}
	}

	// Analyze interaction history to find preferred categories
	summary.PreferredMealCategories = s.analyzePreferredMealCategories(userID)
	summary.PreferredIngredientCats = s.analyzePreferredIngredientCategories(userID)

	return summary, nil
}

func (s *PersonalizedRandomService) analyzePreferredMealCategories(userID string) []string {
	// Get user's top interacted dishes
	topDishes, err := s.interactionService.GetTopInteractedDishes(userID, 30)
	if err != nil || len(topDishes) == 0 {
		return []string{}
	}

	// Count meal categories from top dishes
	categoryCount := make(map[string]int)
	
	for _, slug := range topDishes {
		filter := bson.M{"slug": slug, "deleted": false}
		var dish model.Dish
		s.dishService.Collection().FindOne(context.TODO(), filter).Decode(&dish)
		
		if dish.MealCategories != nil {
			for _, cat := range dish.MealCategories {
				if cat != nil {
					categoryCount[*cat]++
				}
			}
		}
	}

	// Get top 5 categories
	type catScore struct {
		name  string
		count int
	}
	
	categories := make([]catScore, 0, len(categoryCount))
	for name, count := range categoryCount {
		categories = append(categories, catScore{name, count})
	}

	// Sort by count
	for i := 0; i < len(categories); i++ {
		for j := i + 1; j < len(categories); j++ {
			if categories[j].count > categories[i].count {
				categories[i], categories[j] = categories[j], categories[i]
			}
		}
	}

	result := make([]string, 0, 5)
	for i := 0; i < len(categories) && i < 5; i++ {
		result = append(result, categories[i].name)
	}

	return result
}

func (s *PersonalizedRandomService) analyzePreferredIngredientCategories(userID string) []string {
	// Similar logic to meal categories
	topDishes, err := s.interactionService.GetTopInteractedDishes(userID, 30)
	if err != nil || len(topDishes) == 0 {
		return []string{}
	}

	categoryCount := make(map[string]int)
	
	for _, slug := range topDishes {
		filter := bson.M{"slug": slug, "deleted": false}
		var dish model.Dish
		s.dishService.Collection().FindOne(context.TODO(), filter).Decode(&dish)
		
		if dish.IngredientCategories != nil {
			for _, cat := range dish.IngredientCategories {
				if cat != nil {
					categoryCount[*cat]++
				}
			}
		}
	}

	type catScore struct {
		name  string
		count int
	}
	
	categories := make([]catScore, 0, len(categoryCount))
	for name, count := range categoryCount {
		categories = append(categories, catScore{name, count})
	}

	for i := 0; i < len(categories); i++ {
		for j := i + 1; j < len(categories); j++ {
			if categories[j].count > categories[i].count {
				categories[i], categories[j] = categories[j], categories[i]
			}
		}
	}

	result := make([]string, 0, 5)
	for i := 0; i < len(categories) && i < 5; i++ {
		result = append(result, categories[i].name)
	}

	return result
}
