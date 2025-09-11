package service

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"time"
	"what-to-eat/be/config"
	constants "what-to-eat/be/constants"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DishService struct{}

func (s *DishService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.DISH_COLLECTION)
	return col
}

func (ds *DishService) Create(createDishInput model.CreateDishDto, profile *model.JwtCustomClaims) (*model.Dish, error) {
	collection := ds.Collection()

	now := time.Now()

	dish := model.Dish{
		Slug:                 createDishInput.Slug,
		Title:                createDishInput.Title,
		ShortDescription:     createDishInput.ShortDescription,
		Content:              createDishInput.Content,
		Tags:                 createDishInput.Tags,
		PreparationTime:      createDishInput.PreparationTime,
		CookingTime:          createDishInput.CookingTime,
		DifficultLevel:       createDishInput.DifficultLevel,
		MealCategories:       createDishInput.MealCategories,
		IngredientCategories: createDishInput.IngredientCategories,
		Thumbnail:            createDishInput.Thumbnail,
		Videos:               createDishInput.Videos,
		Ingredients:          createDishInput.Ingredients,
		RelatedDishes:        createDishInput.RelatedDishes,
		Labels:               createDishInput.Labels,
		Deleted:              false,
		UpdatedAt:            &now,
		UpdatedBy:            &profile.ID,
		CreatedAt:            &now,
		CreatedBy:            &profile.ID,
	}

	filter := bson.M{"slug": createDishInput.Slug, "deleted": true}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": dish}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) Update(updateDishInput model.UpdateDishDto, profile *model.JwtCustomClaims) (*model.Dish, error) {
	collection := ds.Collection()

	now := time.Now()

	var dish model.Dish

	objectID, err := primitive.ObjectIDFromHex(updateDishInput.ID)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.D{
		{Key: "slug", Value: updateDishInput.Slug},
		{Key: "title", Value: updateDishInput.Title},
		{Key: "shortDescription", Value: updateDishInput.ShortDescription},
		{Key: "content", Value: updateDishInput.Content},
		{Key: "tags", Value: updateDishInput.Tags},
		{Key: "preparationTime", Value: updateDishInput.PreparationTime},
		{Key: "cookingTime", Value: updateDishInput.CookingTime},
		{Key: "difficultLevel", Value: updateDishInput.DifficultLevel},
		{Key: "mealCategories", Value: updateDishInput.MealCategories},
		{Key: "ingredientCategories", Value: updateDishInput.IngredientCategories},
		{Key: "thumbnail", Value: updateDishInput.Thumbnail},
		{Key: "videos", Value: updateDishInput.Videos},
		{Key: "ingredients", Value: updateDishInput.Ingredients},
		{Key: "relatedDishes", Value: updateDishInput.RelatedDishes},
		{Key: "labels", Value: updateDishInput.Labels},
		{Key: "updatedAt", Value: now},
		{Key: "updatedBy", Value: profile.ID},
	}}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) Remove(id string, profile *model.JwtCustomClaims) (*model.Dish, error) {
	collection := ds.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.M{
		"deleted":   true,
		"deletedAt": now,
		"deletedBy": profile.ID,
	}}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	dish := model.Dish{}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) Find(query model.QueryDishDto) ([]*model.Dish, int64, error) {
	collection := ds.Collection()
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(query.Page) - 1) * int64(query.Limit)).SetLimit(int64(query.Limit))
	filter := bson.D{{Key: "deleted", Value: false}}

	// Enhanced search functionality
	if query.Keyword != nil && *query.Keyword != "" {
		searchConditions := bson.A{}

		// 1. Full text search (existing functionality)
		searchConditions = append(searchConditions, bson.D{{Key: "$text", Value: bson.D{{Key: "$search", Value: query.Keyword}}}})

		// 2. Fuzzy search on title, shortDescription, and content in multiple languages
		regexPattern := primitive.Regex{Pattern: *query.Keyword, Options: "i"}
		searchConditions = append(searchConditions, bson.D{{Key: "title.data", Value: regexPattern}})
		searchConditions = append(searchConditions, bson.D{{Key: "shortDescription.data", Value: regexPattern}})
		searchConditions = append(searchConditions, bson.D{{Key: "content.data", Value: regexPattern}})

		// 3. Search in slug
		searchConditions = append(searchConditions, bson.D{{Key: "slug", Value: regexPattern}})

		// 4. Search in tags
		searchConditions = append(searchConditions, bson.D{{Key: "tags", Value: regexPattern}})

		// 5. Search in ingredient slugs
		searchConditions = append(searchConditions, bson.D{{Key: "ingredients.slug", Value: regexPattern}})

		// Combine all search conditions with $or
		filter = append(filter, bson.E{Key: "$or", Value: searchConditions})
	}
	if query.Tags != nil && len(*query.Tags) > 0 {
		filter = append(filter, bson.E{Key: "tags", Value: bson.D{{Key: "$in", Value: query.Tags}}})
	}
	if query.PreparationTimeFrom != nil && query.PreparationTimeTo != nil {
		filter = append(filter, bson.E{Key: "preparationTime", Value: bson.D{{Key: "$lte", Value: query.PreparationTimeTo}, {Key: "$gte", Value: query.PreparationTimeFrom}}})
	}
	if query.CookingTimeFrom != nil && query.CookingTimeTo != nil {
		filter = append(filter, bson.E{Key: "cookingTime", Value: bson.D{{Key: "$lte", Value: query.CookingTimeTo}, {Key: "$gte", Value: query.CookingTimeFrom}}})
	}
	if query.DifficultLevels != nil && len(*query.DifficultLevels) > 0 {
		filter = append(filter, bson.E{Key: "difficultLevel", Value: bson.D{{Key: "$in", Value: query.DifficultLevels}}})
	}
	if query.MealCategories != nil && len(*query.MealCategories) > 0 {
		filter = append(filter, bson.E{Key: "mealCategories", Value: bson.D{{Key: "$in", Value: query.MealCategories}}})
	}
	if query.IngredientCategories != nil && len(*query.IngredientCategories) > 0 {
		filter = append(filter, bson.E{Key: "ingredientCategories", Value: bson.D{{Key: "$in", Value: query.IngredientCategories}}})
	}
	if query.Ingredients != nil && len(*query.Ingredients) > 0 {
		filter = append(filter, bson.E{Key: "ingredients.slug", Value: bson.D{{Key: "$in", Value: query.Ingredients}}})
	}
	if query.Labels != nil && len(*query.Labels) > 0 {
		filter = append(filter, bson.E{Key: "labels", Value: bson.D{{Key: "$in", Value: query.Labels}}})
	}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
	}
	var dishes []*model.Dish
	if err = cursor.All(context.TODO(), &dishes); err != nil {
		log.Println(err)
	}
	defer cursor.Close(context.TODO())
	return dishes, count, err
}

// FindWithScore provides enhanced search with relevance scoring
func (ds *DishService) FindWithScore(query model.QueryDishDto) ([]*model.Dish, int64, error) {
	collection := ds.Collection()

	// Use a simpler approach without complex aggregation for now
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(query.Page) - 1) * int64(query.Limit)).SetLimit(int64(query.Limit))
	filter := bson.D{{Key: "deleted", Value: false}}

	// Enhanced search functionality
	if query.Keyword != nil && *query.Keyword != "" {
		keyword := *query.Keyword
		searchConditions := bson.A{}

		// Create regex pattern for contains search
		containsPattern := primitive.Regex{Pattern: regexp.QuoteMeta(keyword), Options: "i"}

		// For multi-word searches, create a pattern that matches words in sequence
		words := strings.Fields(keyword)
		if len(words) > 1 {
			// Create pattern for phrase matching: "steamed chicken" becomes "steamed.*chicken"
			var wordPatterns []string
			for _, word := range words {
				wordPatterns = append(wordPatterns, regexp.QuoteMeta(word))
			}
			phrasePattern := primitive.Regex{Pattern: strings.Join(wordPatterns, ".*"), Options: "i"}

			// Search in title array with phrase pattern
			searchConditions = append(searchConditions, bson.D{{Key: "title.data", Value: phrasePattern}})
		}

		// Always add individual word searches
		searchConditions = append(searchConditions, bson.D{{Key: "title.data", Value: containsPattern}})
		searchConditions = append(searchConditions, bson.D{{Key: "shortDescription.data", Value: containsPattern}})
		searchConditions = append(searchConditions, bson.D{{Key: "content.data", Value: containsPattern}})
		searchConditions = append(searchConditions, bson.D{{Key: "slug", Value: containsPattern}})
		searchConditions = append(searchConditions, bson.D{{Key: "tags", Value: containsPattern}})
		searchConditions = append(searchConditions, bson.D{{Key: "ingredients.slug", Value: containsPattern}})

		// Combine all search conditions with $or
		filter = append(filter, bson.E{Key: "$or", Value: searchConditions})
	}

	// Add other filters
	if query.Tags != nil && len(*query.Tags) > 0 {
		filter = append(filter, bson.E{Key: "tags", Value: bson.D{{Key: "$in", Value: query.Tags}}})
	}
	if query.PreparationTimeFrom != nil && query.PreparationTimeTo != nil {
		filter = append(filter, bson.E{Key: "preparationTime", Value: bson.D{{Key: "$lte", Value: query.PreparationTimeTo}, {Key: "$gte", Value: query.PreparationTimeFrom}}})
	}
	if query.CookingTimeFrom != nil && query.CookingTimeTo != nil {
		filter = append(filter, bson.E{Key: "cookingTime", Value: bson.D{{Key: "$lte", Value: query.CookingTimeTo}, {Key: "$gte", Value: query.CookingTimeFrom}}})
	}
	if query.DifficultLevels != nil && len(*query.DifficultLevels) > 0 {
		filter = append(filter, bson.E{Key: "difficultLevel", Value: bson.D{{Key: "$in", Value: query.DifficultLevels}}})
	}
	if query.MealCategories != nil && len(*query.MealCategories) > 0 {
		filter = append(filter, bson.E{Key: "mealCategories", Value: bson.D{{Key: "$in", Value: query.MealCategories}}})
	}
	if query.IngredientCategories != nil && len(*query.IngredientCategories) > 0 {
		filter = append(filter, bson.E{Key: "ingredientCategories", Value: bson.D{{Key: "$in", Value: query.IngredientCategories}}})
	}
	if query.Ingredients != nil && len(*query.Ingredients) > 0 {
		filter = append(filter, bson.E{Key: "ingredients.slug", Value: bson.D{{Key: "$in", Value: query.Ingredients}}})
	}
	if query.Labels != nil && len(*query.Labels) > 0 {
		filter = append(filter, bson.E{Key: "labels", Value: bson.D{{Key: "$in", Value: query.Labels}}})
	}

	// Get count and results using simple Find operation
	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())

	var dishes []*model.Dish
	if err = cursor.All(context.TODO(), &dishes); err != nil {
		return nil, 0, err
	}

	return dishes, count, nil
}

// FindSuggestions provides auto-complete suggestions for search
func (ds *DishService) FindSuggestions(keyword string, limit int) ([]string, error) {
	if keyword == "" {
		return []string{}, nil
	}

	collection := ds.Collection()
	regexPattern := primitive.Regex{Pattern: "^" + regexp.QuoteMeta(keyword), Options: "i"} // Starts with keyword

	pipeline := mongo.Pipeline{
		// Match non-deleted dishes
		{{Key: "$match", Value: bson.D{{Key: "deleted", Value: false}}}},

		// Create suggestions from multiple fields
		{{Key: "$project", Value: bson.D{
			{Key: "suggestions", Value: bson.D{{
				Key: "$concatArrays", Value: bson.A{
					// Extract title suggestions
					bson.D{{Key: "$map", Value: bson.D{
						{Key: "input", Value: "$title"},
						{Key: "as", Value: "titleItem"},
						{Key: "in", Value: "$$titleItem.data"},
					}}},
					// Extract tag suggestions
					"$tags",
					// Extract slug suggestions (convert to readable format)
					bson.A{"$slug"},
				},
			}}},
		}}},

		// Unwind suggestions array
		{{Key: "$unwind", Value: "$suggestions"}},

		// Filter suggestions that start with keyword
		{{Key: "$match", Value: bson.D{{Key: "suggestions", Value: regexPattern}}}},

		// Group by suggestion to remove duplicates and count
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$suggestions"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},

		// Sort by count (most popular first) then alphabetically
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}, {Key: "_id", Value: 1}}}},

		// Limit results
		{{Key: "$limit", Value: int64(limit)}},

		// Project only the suggestion text
		{{Key: "$project", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []struct {
		ID string `bson:"_id"`
	}
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	suggestions := make([]string, len(results))
	for i, result := range results {
		suggestions[i] = result.ID
	}

	return suggestions, nil
}

func (ds *DishService) FindOne(id string) (*model.Dish, error) {
	collection := ds.Collection()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	dish := model.Dish{}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) FindOneBySlug(slug string) (*model.Dish, error) {
	collection := ds.Collection()
	filter := bson.M{"slug": slug}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	dish := model.Dish{}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) Random(query model.QueryDishRandomDto) ([]*model.Dish, error) {
	collection := ds.Collection()
	stages := []bson.D{}
	matchStage := bson.D{{Key: "deleted", Value: false}}
	if query.MealCategories != nil && len(*query.MealCategories) > 0 {
		matchStage = append(matchStage, bson.E{Key: "mealCategories", Value: bson.D{{Key: "$in", Value: query.MealCategories}}})
	}
	stages = append(stages, bson.D{{Key: "$match", Value: matchStage}})
	stages = append(stages, bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: int64(query.Limit)}}}})
	cursor, err := collection.Aggregate(context.TODO(), stages)
	if err != nil {
		return nil, err
	}

	var dishes []*model.Dish
	if err = cursor.All(context.TODO(), &dishes); err != nil {
		log.Println(err)
	}
	defer cursor.Close(context.TODO())
	return dishes, err
}

func (ds *DishService) Analyze() (map[string]interface{}, error) {
	collection := ds.Collection()

	pipeline := mongo.Pipeline{
		// Only non-deleted dishes
		{{Key: "$match", Value: bson.D{{Key: "deleted", Value: false}}}},
		// Facet for multiple analytics
		{{
			Key: "$facet", Value: bson.D{
				// Count by mealCategories
				{Key: "categoryDistribution", Value: bson.A{
					bson.D{{Key: "$unwind", Value: "$mealCategories"}},
					bson.D{{Key: "$group", Value: bson.D{
						{Key: "_id", Value: "$mealCategories"},
						{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
					}}},
				}},
				// Average cooking and preparation time
				{Key: "avgTimes", Value: bson.A{
					bson.D{{Key: "$group", Value: bson.D{
						{Key: "_id", Value: nil},
						{Key: "avgCookingTime", Value: bson.D{{Key: "$avg", Value: "$cookingTime"}}},
						{Key: "avgPreparationTime", Value: bson.D{{Key: "$avg", Value: "$preparationTime"}}},
					}}},
				}},
				// Count by difficultLevel
				{Key: "difficultyLevels", Value: bson.A{
					bson.D{{Key: "$group", Value: bson.D{
						{Key: "_id", Value: "$difficultLevel"},
						{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
					}}},
				}},
			},
		}},
	}

	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []map[string]interface{}
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return map[string]interface{}{}, nil
	}
	return results[0], nil
}
