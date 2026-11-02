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

type DishService struct {
	col CollectionInterface
}

func (s *DishService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.DISH_COLLECTION)
	return col
}

// NewDishService creates a DishService with a pre-configured CollectionInterface (useful for testing).
func NewDishService(col CollectionInterface) *DishService {
	return &DishService{col: col}
}

// getCol returns the injected collection or wraps the real MongoDB collection.
func (ds *DishService) getCol() CollectionInterface {
	if ds.col != nil {
		return ds.col
	}
	return NewMongoCollectionAdapter(ds.Collection())
}

func (ds *DishService) Create(createDishInput model.CreateDishDto, profile *model.JwtCustomClaims) (*model.Dish, error) {
	collection := ds.getCol()

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
	collection := ds.getCol()

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
	collection := ds.getCol()
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
	collection := ds.getCol()
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

// FindWithScore provides Google-like search with advanced relevance scoring
func (ds *DishService) FindWithScore(query model.QueryDishDto) ([]*model.Dish, int64, error) {
	collection := ds.getCol()

	// If no search keyword, fall back to regular Find with filters
	if query.Keyword == nil || *query.Keyword == "" {
		return ds.Find(query)
	}

	keyword := strings.TrimSpace(*query.Keyword)
	words := strings.Fields(strings.ToLower(keyword))

	// Use simpler approach with MongoDB text search and regex
	pipeline := ds.buildSimpleSearchPipeline(keyword, words, query)

	// Get total count first (without pagination)
	countPipeline := make(mongo.Pipeline, len(pipeline)-2) // Remove skip and limit
	copy(countPipeline, pipeline[:len(pipeline)-2])
	countPipeline = append(countPipeline, bson.D{{Key: "$count", Value: "total"}})

	countCursor, err := collection.Aggregate(context.TODO(), countPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer countCursor.Close(context.TODO())

	var countResult []struct {
		Total int64 `bson:"total"`
	}
	if err = countCursor.All(context.TODO(), &countResult); err != nil {
		return nil, 0, err
	}

	var totalCount int64 = 0
	if len(countResult) > 0 {
		totalCount = countResult[0].Total
	}

	// Get paginated results
	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())

	var results []struct {
		Document model.Dish `bson:"document"`
		Score    float64    `bson:"score"`
	}
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, 0, err
	}

	// Extract dishes from results
	dishes := make([]*model.Dish, len(results))
	for i, result := range results {
		dishes[i] = &result.Document
	}

	return dishes, totalCount, nil
} // buildSearchPipeline creates a comprehensive search pipeline with scoring
func (ds *DishService) buildSearchPipeline(keyword string, words []string, query model.QueryDishDto) mongo.Pipeline {
	pipeline := mongo.Pipeline{}

	// Build match conditions as an $and array to properly combine all filters
	andConditions := bson.A{
		bson.D{{Key: "deleted", Value: false}},
	}

	// Add other filters first
	if query.Tags != nil && len(*query.Tags) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "tags", Value: bson.D{{Key: "$in", Value: query.Tags}}}})
	}
	if query.PreparationTimeFrom != nil && query.PreparationTimeTo != nil {
		// Include dishes with null preparationTime OR within the specified range
		preparationTimeFilter := bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "preparationTime", Value: nil}},
			bson.D{{Key: "preparationTime", Value: bson.D{{Key: "$lte", Value: query.PreparationTimeTo}, {Key: "$gte", Value: query.PreparationTimeFrom}}}},
		}}}
		andConditions = append(andConditions, preparationTimeFilter)
	}
	if query.CookingTimeFrom != nil && query.CookingTimeTo != nil {
		// Include dishes with null cookingTime OR within the specified range
		cookingTimeFilter := bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "cookingTime", Value: nil}},
			bson.D{{Key: "cookingTime", Value: bson.D{{Key: "$lte", Value: query.CookingTimeTo}, {Key: "$gte", Value: query.CookingTimeFrom}}}},
		}}}
		andConditions = append(andConditions, cookingTimeFilter)
	}
	if query.DifficultLevels != nil && len(*query.DifficultLevels) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "difficultLevel", Value: bson.D{{Key: "$in", Value: query.DifficultLevels}}}})
	}
	if query.MealCategories != nil && len(*query.MealCategories) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "mealCategories", Value: bson.D{{Key: "$in", Value: query.MealCategories}}}})
	}
	if query.IngredientCategories != nil && len(*query.IngredientCategories) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "ingredientCategories", Value: bson.D{{Key: "$in", Value: query.IngredientCategories}}}})
	}
	if query.Ingredients != nil && len(*query.Ingredients) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "ingredients.slug", Value: bson.D{{Key: "$in", Value: query.Ingredients}}}})
	}
	if query.Labels != nil && len(*query.Labels) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "labels", Value: bson.D{{Key: "$in", Value: query.Labels}}}})
	}

	// Create the match stage with $and combining all conditions
	matchStage := bson.D{{Key: "$and", Value: andConditions}}
	pipeline = append(pipeline, bson.D{{Key: "$match", Value: matchStage}})

	// Add search scoring stage
	pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: bson.D{
		{Key: "score", Value: ds.buildScoreExpression(keyword, words)},
	}}})

	// Filter out documents with score 0 (no matches)
	pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "score", Value: bson.D{{Key: "$gt", Value: 0}}}}}})

	// Sort by relevance score (descending), then by creation date
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{
		{Key: "score", Value: -1},
		{Key: "createdAt", Value: -1},
	}}})

	// Add pagination
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: (int64(query.Page) - 1) * int64(query.Limit)}})
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: int64(query.Limit)}})

	// Project the final structure
	pipeline = append(pipeline, bson.D{{Key: "$project", Value: bson.D{
		{Key: "document", Value: "$$ROOT"},
		{Key: "score", Value: 1},
	}}})

	return pipeline
}

// buildScoreExpression creates a comprehensive scoring expression
func (ds *DishService) buildScoreExpression(keyword string, words []string) interface{} {
	keywordLower := strings.ToLower(keyword)

	// Create regex patterns for different match types
	exactPattern := primitive.Regex{Pattern: "^" + regexp.QuoteMeta(keywordLower) + "$", Options: "i"}
	startsWithPattern := primitive.Regex{Pattern: "^" + regexp.QuoteMeta(keywordLower), Options: "i"}
	containsPattern := primitive.Regex{Pattern: regexp.QuoteMeta(keywordLower), Options: "i"}

	// Create phrase pattern for multi-word searches
	var phrasePattern primitive.Regex
	if len(words) > 1 {
		wordPatterns := make([]string, len(words))
		for i, word := range words {
			wordPatterns[i] = regexp.QuoteMeta(word)
		}
		phrasePattern = primitive.Regex{Pattern: strings.Join(wordPatterns, ".*"), Options: "i"}
	}

	// Score components
	scoreComponents := bson.A{}

	// Helper function to create field scoring
	createFieldScore := func(fieldPath string, exactWeight, startsWithWeight, containsWeight, phraseWeight float64) bson.D {
		fieldScores := bson.A{}

		// Exact match (highest score)
		fieldScores = append(fieldScores, bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$regexMatch", Value: bson.D{
				{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$" + fieldPath}}},
				{Key: "regex", Value: exactPattern},
			}}},
			exactWeight,
			0,
		}}})

		// Starts with match
		fieldScores = append(fieldScores, bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$regexMatch", Value: bson.D{
				{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$" + fieldPath}}},
				{Key: "regex", Value: startsWithPattern},
			}}},
			startsWithWeight,
			0,
		}}})

		// Contains match
		fieldScores = append(fieldScores, bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$regexMatch", Value: bson.D{
				{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$" + fieldPath}}},
				{Key: "regex", Value: containsPattern},
			}}},
			containsWeight,
			0,
		}}})

		// Phrase match for multi-word queries
		if len(words) > 1 {
			fieldScores = append(fieldScores, bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$" + fieldPath}}},
					{Key: "regex", Value: phrasePattern},
				}}},
				phraseWeight,
				0,
			}}})
		}

		return bson.D{{Key: "$max", Value: fieldScores}}
	} // Score title fields (highest priority) - using simple array scoring
	var titleScoreConditions bson.A
	titleScoreConditions = append(titleScoreConditions,
		// Exact match
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$regexMatch", Value: bson.D{
				{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
				{Key: "regex", Value: exactPattern},
			}}},
			100,
			0,
		}}},
		// Starts with
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$regexMatch", Value: bson.D{
				{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
				{Key: "regex", Value: startsWithPattern},
			}}},
			80,
			0,
		}}},
		// Contains
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$regexMatch", Value: bson.D{
				{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
				{Key: "regex", Value: containsPattern},
			}}},
			60,
			0,
		}}},
	)

	// Add phrase matching for multi-word queries
	if len(words) > 1 {
		titleScoreConditions = append(titleScoreConditions,
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
					{Key: "regex", Value: phrasePattern},
				}}},
				70,
				0,
			}}},
		)
	}

	titleScore := bson.D{{Key: "$sum", Value: bson.D{{Key: "$map", Value: bson.D{
		{Key: "input", Value: "$title"},
		{Key: "as", Value: "item"},
		{Key: "in", Value: bson.D{{Key: "$max", Value: titleScoreConditions}}},
	}}}}}

	// Score short description (high priority)
	descScore := bson.D{{Key: "$sum", Value: bson.D{{Key: "$map", Value: bson.D{
		{Key: "input", Value: "$shortDescription"},
		{Key: "as", Value: "item"},
		{Key: "in", Value: bson.D{{Key: "$max", Value: bson.A{
			// Exact match
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
					{Key: "regex", Value: exactPattern},
				}}},
				80,
				0,
			}}},
			// Starts with
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
					{Key: "regex", Value: startsWithPattern},
				}}},
				60,
				0,
			}}},
			// Contains
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
					{Key: "regex", Value: containsPattern},
				}}},
				40,
				0,
			}}},
		}}}},
	}}}}}

	// Score content (medium priority)
	contentScore := bson.D{{Key: "$sum", Value: bson.D{{Key: "$map", Value: bson.D{
		{Key: "input", Value: "$content"},
		{Key: "as", Value: "item"},
		{Key: "in", Value: bson.D{{Key: "$max", Value: bson.A{
			// Exact match
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
					{Key: "regex", Value: exactPattern},
				}}},
				60,
				0,
			}}},
			// Starts with
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
					{Key: "regex", Value: startsWithPattern},
				}}},
				40,
				0,
			}}},
			// Contains
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
					{Key: "regex", Value: containsPattern},
				}}},
				20,
				0,
			}}},
		}}}},
	}}}}}

	// Score slug (high priority for SEO)
	slugScore := createFieldScore("slug", 90, 70, 50, 60)

	// Score tags (medium-high priority)
	tagScore := bson.D{{Key: "$sum", Value: bson.D{{Key: "$map", Value: bson.D{
		{Key: "input", Value: "$tags"},
		{Key: "as", Value: "tag"},
		{Key: "in", Value: bson.D{{Key: "$max", Value: bson.A{
			// Exact match
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$tag"}}},
					{Key: "regex", Value: exactPattern},
				}}},
				70,
				0,
			}}},
			// Starts with
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$tag"}}},
					{Key: "regex", Value: startsWithPattern},
				}}},
				50,
				0,
			}}},
			// Contains
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$tag"}}},
					{Key: "regex", Value: containsPattern},
				}}},
				30,
				0,
			}}},
		}}}},
	}}}}}

	// Score ingredient slugs (medium priority)
	ingredientScore := bson.D{{Key: "$sum", Value: bson.D{{Key: "$map", Value: bson.D{
		{Key: "input", Value: "$ingredients"},
		{Key: "as", Value: "ingredient"},
		{Key: "in", Value: bson.D{{Key: "$max", Value: bson.A{
			// Exact match
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$ingredient.slug"}}},
					{Key: "regex", Value: exactPattern},
				}}},
				50,
				0,
			}}},
			// Starts with
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$ingredient.slug"}}},
					{Key: "regex", Value: startsWithPattern},
				}}},
				30,
				0,
			}}},
			// Contains
			bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$regexMatch", Value: bson.D{
					{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$ingredient.slug"}}},
					{Key: "regex", Value: containsPattern},
				}}},
				15,
				0,
			}}},
		}}}},
	}}}}}

	// Individual word scoring for multi-word queries
	if len(words) > 1 {
		for _, word := range words {
			wordPattern := primitive.Regex{Pattern: regexp.QuoteMeta(word), Options: "i"}

			wordScore := bson.D{{Key: "$add", Value: bson.A{
				// Title word match
				bson.D{{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$anyElementTrue", Value: bson.D{{Key: "$map", Value: bson.D{
						{Key: "input", Value: "$title"},
						{Key: "as", Value: "item"},
						{Key: "in", Value: bson.D{{Key: "$regexMatch", Value: bson.D{
							{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
							{Key: "regex", Value: wordPattern},
						}}}},
					}}}}},
					15,
					0,
				}}},
				// Description word match
				bson.D{{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$anyElementTrue", Value: bson.D{{Key: "$map", Value: bson.D{
						{Key: "input", Value: "$shortDescription"},
						{Key: "as", Value: "item"},
						{Key: "in", Value: bson.D{{Key: "$regexMatch", Value: bson.D{
							{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item.data"}}},
							{Key: "regex", Value: wordPattern},
						}}}},
					}}}}},
					10,
					0,
				}}},
				// Tag word match
				bson.D{{Key: "$cond", Value: bson.A{
					bson.D{{Key: "$anyElementTrue", Value: bson.D{{Key: "$map", Value: bson.D{
						{Key: "input", Value: "$tags"},
						{Key: "as", Value: "item"},
						{Key: "in", Value: bson.D{{Key: "$regexMatch", Value: bson.D{
							{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$item"}}},
							{Key: "regex", Value: wordPattern},
						}}}},
					}}}}},
					8,
					0,
				}}},
			}}}

			scoreComponents = append(scoreComponents, wordScore)
		}
	} // Combine all score components
	scoreComponents = append(scoreComponents, titleScore, descScore, contentScore, slugScore, tagScore, ingredientScore)

	// Add popularity boost based on creation date (newer dishes get slight boost)
	popularityBoost := bson.D{{Key: "$cond", Value: bson.A{
		bson.D{{Key: "$gte", Value: bson.A{
			"$createdAt",
			bson.D{{Key: "$dateSubtract", Value: bson.D{
				{Key: "startDate", Value: "$$NOW"},
				{Key: "unit", Value: "month"},
				{Key: "amount", Value: 3},
			}}},
		}}},
		5, // Boost for dishes created in last 3 months
		0,
	}}}

	scoreComponents = append(scoreComponents, popularityBoost)

	return bson.D{{Key: "$add", Value: scoreComponents}}
}

// buildSimpleSearchPipeline creates a simplified search pipeline
func (ds *DishService) buildSimpleSearchPipeline(keyword string, words []string, query model.QueryDishDto) mongo.Pipeline {
	pipeline := mongo.Pipeline{}

	// Build match conditions as an $and array to properly combine all filters
	andConditions := bson.A{
		bson.D{{Key: "deleted", Value: false}},
	}

	// Add keyword search using simple regex on multiple fields
	if keyword != "" {
		containsPattern := primitive.Regex{Pattern: regexp.QuoteMeta(keyword), Options: "i"}

		searchConditions := bson.A{
			// Search in title array
			bson.D{{Key: "title.data", Value: containsPattern}},
			// Search in short description array
			bson.D{{Key: "shortDescription.data", Value: containsPattern}},
			// Search in content array
			bson.D{{Key: "content.data", Value: containsPattern}},
			// Search in slug
			bson.D{{Key: "slug", Value: containsPattern}},
			// Search in tags
			bson.D{{Key: "tags", Value: containsPattern}},
			// Search in ingredient slugs
			bson.D{{Key: "ingredients.slug", Value: containsPattern}},
		}

		andConditions = append(andConditions, bson.D{{Key: "$or", Value: searchConditions}})
	}

	// Add other filters
	if query.Tags != nil && len(*query.Tags) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "tags", Value: bson.D{{Key: "$in", Value: query.Tags}}}})
	}
	if query.PreparationTimeFrom != nil && query.PreparationTimeTo != nil {
		// Include dishes with null preparationTime OR within the specified range
		preparationTimeFilter := bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "preparationTime", Value: nil}},
			bson.D{{Key: "preparationTime", Value: bson.D{{Key: "$lte", Value: query.PreparationTimeTo}, {Key: "$gte", Value: query.PreparationTimeFrom}}}},
		}}}
		andConditions = append(andConditions, preparationTimeFilter)
	}
	if query.CookingTimeFrom != nil && query.CookingTimeTo != nil {
		// Include dishes with null cookingTime OR within the specified range
		cookingTimeFilter := bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "cookingTime", Value: nil}},
			bson.D{{Key: "cookingTime", Value: bson.D{{Key: "$lte", Value: query.CookingTimeTo}, {Key: "$gte", Value: query.CookingTimeFrom}}}},
		}}}
		andConditions = append(andConditions, cookingTimeFilter)
	}
	if query.DifficultLevels != nil && len(*query.DifficultLevels) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "difficultLevel", Value: bson.D{{Key: "$in", Value: query.DifficultLevels}}}})
	}
	if query.MealCategories != nil && len(*query.MealCategories) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "mealCategories", Value: bson.D{{Key: "$in", Value: query.MealCategories}}}})
	}
	if query.IngredientCategories != nil && len(*query.IngredientCategories) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "ingredientCategories", Value: bson.D{{Key: "$in", Value: query.IngredientCategories}}}})
	}
	if query.Ingredients != nil && len(*query.Ingredients) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "ingredients.slug", Value: bson.D{{Key: "$in", Value: query.Ingredients}}}})
	}
	if query.Labels != nil && len(*query.Labels) > 0 {
		andConditions = append(andConditions, bson.D{{Key: "labels", Value: bson.D{{Key: "$in", Value: query.Labels}}}})
	}

	// Create the match stage with $and combining all conditions
	matchStage := bson.D{{Key: "$and", Value: andConditions}}

	pipeline = append(pipeline, bson.D{{Key: "$match", Value: matchStage}})

	// Add simple scoring based on field matches
	scoreExpression := ds.buildSimpleScoreExpression(keyword)
	pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: bson.D{
		{Key: "score", Value: scoreExpression},
	}}})

	// Sort by score descending, then by creation date
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{
		{Key: "score", Value: -1},
		{Key: "createdAt", Value: -1},
	}}})

	// Add pagination
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: (int64(query.Page) - 1) * int64(query.Limit)}})
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: int64(query.Limit)}})

	// Project the final structure
	pipeline = append(pipeline, bson.D{{Key: "$project", Value: bson.D{
		{Key: "document", Value: "$$ROOT"},
		{Key: "score", Value: 1},
	}}})

	return pipeline
}

// buildSimpleScoreExpression creates a simplified scoring expression
func (ds *DishService) buildSimpleScoreExpression(keyword string) interface{} {
	if keyword == "" {
		return 1 // Default score
	}

	keywordLower := strings.ToLower(keyword)
	containsPattern := primitive.Regex{Pattern: regexp.QuoteMeta(keywordLower), Options: "i"}
	exactPattern := primitive.Regex{Pattern: "^" + regexp.QuoteMeta(keywordLower) + "$", Options: "i"}
	startsWithPattern := primitive.Regex{Pattern: "^" + regexp.QuoteMeta(keywordLower), Options: "i"}

	// Calculate score based on different field matches
	scoreComponents := bson.A{
		// Title exact match (highest priority)
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$gt", Value: bson.A{
				bson.D{{Key: "$size", Value: bson.D{{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$title"},
					{Key: "cond", Value: bson.D{{Key: "$regexMatch", Value: bson.D{
						{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$this.data"}}},
						{Key: "regex", Value: exactPattern},
					}}}},
				}}}}},
				0,
			}}},
			100,
			0,
		}}},

		// Title starts with
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$gt", Value: bson.A{
				bson.D{{Key: "$size", Value: bson.D{{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$title"},
					{Key: "cond", Value: bson.D{{Key: "$regexMatch", Value: bson.D{
						{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$this.data"}}},
						{Key: "regex", Value: startsWithPattern},
					}}}},
				}}}}},
				0,
			}}},
			80,
			0,
		}}},

		// Title contains
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$gt", Value: bson.A{
				bson.D{{Key: "$size", Value: bson.D{{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$title"},
					{Key: "cond", Value: bson.D{{Key: "$regexMatch", Value: bson.D{
						{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$this.data"}}},
						{Key: "regex", Value: containsPattern},
					}}}},
				}}}}},
				0,
			}}},
			60,
			0,
		}}},

		// Slug exact match
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$regexMatch", Value: bson.D{
				{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$slug"}}},
				{Key: "regex", Value: exactPattern},
			}}},
			90,
			0,
		}}},

		// Slug contains
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$regexMatch", Value: bson.D{
				{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$slug"}}},
				{Key: "regex", Value: containsPattern},
			}}},
			50,
			0,
		}}},

		// Short description contains
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$gt", Value: bson.A{
				bson.D{{Key: "$size", Value: bson.D{{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$shortDescription"},
					{Key: "cond", Value: bson.D{{Key: "$regexMatch", Value: bson.D{
						{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$this.data"}}},
						{Key: "regex", Value: containsPattern},
					}}}},
				}}}}},
				0,
			}}},
			40,
			0,
		}}},

		// Tags match
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$gt", Value: bson.A{
				bson.D{{Key: "$size", Value: bson.D{{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$tags"},
					{Key: "cond", Value: bson.D{{Key: "$regexMatch", Value: bson.D{
						{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$this"}}},
						{Key: "regex", Value: containsPattern},
					}}}},
				}}}}},
				0,
			}}},
			30,
			0,
		}}},

		// Ingredient slug match
		bson.D{{Key: "$cond", Value: bson.A{
			bson.D{{Key: "$gt", Value: bson.A{
				bson.D{{Key: "$size", Value: bson.D{{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$ingredients"},
					{Key: "cond", Value: bson.D{{Key: "$regexMatch", Value: bson.D{
						{Key: "input", Value: bson.D{{Key: "$toLower", Value: "$$this.slug"}}},
						{Key: "regex", Value: containsPattern},
					}}}},
				}}}}},
				0,
			}}},
			25,
			0,
		}}},
	}

	return bson.D{{Key: "$add", Value: scoreComponents}}
}

// FindWithFuzzyScore provides fuzzy search with typo tolerance and improved ranking
func (ds *DishService) FindWithFuzzyScore(query model.QueryDishDto) ([]*model.Dish, int64, error) {
	if query.Keyword == nil || *query.Keyword == "" {
		return ds.Find(query)
	}

	keyword := strings.TrimSpace(*query.Keyword)

	// Try exact search first - prioritize exact matches
	exactResults, exactCount, err := ds.FindWithScore(query)
	if err != nil {
		return nil, 0, err
	}

	// For multi-word phrases with exact matches, prioritize them heavily
	words := strings.Fields(strings.ToLower(keyword))
	if len(words) > 1 && exactCount > 0 {
		// Multi-word exact matches should be highly trusted - return them first
		// Add debug log to verify this path is taken
		log.Printf("DEBUG: Found %d exact matches for multi-word phrase: %s", exactCount, keyword)

		// But also add some fuzzy variants for better coverage
		fuzzyVariants := ds.generateFuzzyVariants(keyword)
		seenDishes := make(map[string]*model.Dish)
		var combinedResults []*model.Dish

		// Add exact results first (highest priority)
		for _, dish := range exactResults {
			seenDishes[dish.Slug] = dish
			combinedResults = append(combinedResults, dish)
			log.Printf("DEBUG: Adding exact match: %s", dish.Slug)
		}

		// Add a few fuzzy results for coverage, but limited
		for i, variant := range fuzzyVariants {
			if i >= 3 { // Limit to first 3 variants to avoid noise
				break
			}
			if variant == keyword {
				continue
			}
			log.Printf("DEBUG: Trying fuzzy variant: %s", variant)

			fuzzyQuery := query
			fuzzyQuery.Keyword = &variant

			variantDishes, _, err := ds.FindWithScore(fuzzyQuery)
			if err != nil {
				continue
			}

			// Add only a few additional results
			added := 0
			for _, dish := range variantDishes {
				if _, seen := seenDishes[dish.Slug]; !seen && added < 2 {
					seenDishes[dish.Slug] = dish
					combinedResults = append(combinedResults, dish)
					added++
					log.Printf("DEBUG: Adding fuzzy result: %s", dish.Slug)
				}
			}
		}

		log.Printf("DEBUG: Returning %d combined results", len(combinedResults))
		return combinedResults, int64(len(combinedResults)), nil
	}

	// If single word or no exact matches, continue with fuzzy search
	fuzzyVariants := ds.generateFuzzyVariants(keyword)

	// Use map to collect all results and avoid duplicates
	seenDishes := make(map[string]*model.Dish)
	var allResults []*model.Dish

	// Add exact results first (highest priority)
	for _, dish := range exactResults {
		seenDishes[dish.Slug] = dish
		allResults = append(allResults, dish)
	}

	// Try fuzzy variants with lower priority
	for _, variant := range fuzzyVariants {
		if variant == keyword {
			continue // Skip original keyword as we already searched it
		}

		fuzzyQuery := query
		fuzzyQuery.Keyword = &variant

		variantDishes, _, err := ds.FindWithScore(fuzzyQuery)
		if err != nil {
			continue
		}

		// Add variant results that we haven't seen before
		for _, dish := range variantDishes {
			if _, seen := seenDishes[dish.Slug]; !seen {
				seenDishes[dish.Slug] = dish
				allResults = append(allResults, dish)
			}
		}

		// Limit total results to avoid too many irrelevant matches
		if len(allResults) >= 20 {
			break
		}
	}

	return allResults, int64(len(allResults)), nil
}

// generateFuzzyVariants creates variations of the search term to handle typos
func (ds *DishService) generateFuzzyVariants(keyword string) []string {
	variants := []string{keyword}
	words := strings.Fields(strings.ToLower(keyword))

	// Common cooking term corrections with Vietnamese support
	corrections := map[string]string{
		"chiken":    "chicken",
		"chickne":   "chicken",
		"chicke":    "chicken",
		"chikken":   "chicken",
		"tomatoe":   "tomato",
		"tomatos":   "tomatoes",
		"potatoe":   "potato",
		"potatos":   "potatoes",
		"onoin":     "onion",
		"oinon":     "onion",
		"carott":    "carrot",
		"carot":     "carrot",
		"spagetti":  "spaghetti",
		"spageti":   "spaghetti",
		"brocolli":  "broccoli",
		"brocoli":   "broccoli",
		"mashroom":  "mushroom",
		"mushrrom":  "mushroom",
		"musroom":   "mushroom",
		"beaf":      "beef",
		"beff":      "beef",
		"prok":      "pork",
		"porc":      "pork",
		"salman":    "salmon",
		"samon":     "salmon",
		"shrimps":   "shrimp",
		"shremp":    "shrimp",
		"chease":    "cheese",
		"chees":     "cheese",
		"cheeze":    "cheese",
		"letuce":    "lettuce",
		"lettuze":   "lettuce",
		"cumcumber": "cucumber",
		"cucmber":   "cucumber",
		"avacado":   "avocado",
		// Vietnamese cooking term corrections
		"ban":   "bánh",
		"banh":  "bánh",
		"bao":   "bánh bao",
		"chien": "chiên",
		"nuong": "nướng",
		"ruoc":  "ruốc",
		"bak":   "bánh",
	}

	// Apply corrections to each word
	correctedWords := make([]string, len(words))
	hasCorrection := false

	for i, word := range words {
		if correction, exists := corrections[word]; exists {
			correctedWords[i] = correction
			hasCorrection = true
		} else {
			correctedWords[i] = word
		}
	}

	if hasCorrection {
		variants = append(variants, strings.Join(correctedWords, " "))
	}

	// For multi-word phrases, be more conservative with partial matches
	// to avoid ranking unrelated dishes higher than exact matches
	if len(words) > 1 {
		// For multi-word phrases, only add very conservative variations
		// Add word-by-word search only for the main terms
		for _, word := range words {
			if len(word) > 3 {
				variants = append(variants, word)
			}
		}
	} else {
		// For single words, add more aggressive partial matches
		for _, word := range words {
			if len(word) > 4 {
				// Add prefix matches
				variants = append(variants, word[:len(word)-1]) // Remove last character
				if len(word) > 5 {
					variants = append(variants, word[:len(word)-2]) // Remove last 2 characters
				}
			}
		}
	}

	// Remove duplicates
	uniqueVariants := make([]string, 0)
	seen := make(map[string]bool)

	for _, variant := range variants {
		if !seen[variant] && variant != "" {
			uniqueVariants = append(uniqueVariants, variant)
			seen[variant] = true
		}
	}

	return uniqueVariants
}

// FindSmart combines multiple search strategies for best results
func (ds *DishService) FindSmart(query model.QueryDishDto) ([]*model.Dish, int64, error) {
	if query.Keyword == nil || *query.Keyword == "" {
		return ds.Find(query)
	}

	keyword := strings.TrimSpace(*query.Keyword)

	// Strategy 1: Try exact phrase search first
	results, count, err := ds.FindWithScore(query)
	if err != nil {
		return nil, 0, err
	}

	// If we have good results (>= 5), return them
	if count >= 5 {
		return results, count, nil
	}

	// Strategy 2: If few results, try individual word search
	words := strings.Fields(keyword)
	if len(words) > 1 {
		wordResults := make(map[string]*model.Dish)
		totalWordCount := int64(0)

		for _, word := range words {
			if len(word) < 3 {
				continue // Skip very short words
			}

			wordQuery := query
			wordQuery.Keyword = &word

			wordDishes, _, err := ds.FindWithScore(wordQuery)
			if err != nil {
				continue
			}

			// Add unique dishes
			for _, dish := range wordDishes {
				if _, exists := wordResults[dish.ID]; !exists {
					wordResults[dish.ID] = dish
					totalWordCount++
				}
			}
		}

		// If word-based search gives us more results, use it
		if totalWordCount > count {
			wordDishes := make([]*model.Dish, 0, len(wordResults))
			for _, dish := range wordResults {
				wordDishes = append(wordDishes, dish)
			}
			return wordDishes, totalWordCount, nil
		}
	}

	// Strategy 3: Try fuzzy search as last resort
	if count < 3 {
		return ds.FindWithFuzzyScore(query)
	}

	return results, count, nil
}

// FindSuggestions provides auto-complete suggestions for search
func (ds *DishService) FindSuggestions(keyword string, limit int) ([]string, error) {
	if keyword == "" {
		return []string{}, nil
	}

	collection := ds.getCol()
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
	collection := ds.getCol()

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
	collection := ds.getCol()
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
	collection := ds.getCol()
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
	collection := ds.getCol()

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
