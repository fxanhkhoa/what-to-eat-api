package service

import (
	"context"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DishPopularityService struct{}

func NewDishPopularityService() *DishPopularityService {
	return &DishPopularityService{}
}

func (s *DishPopularityService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.DISH_POPULARITY_COLLECTION)
	return col
}

func (s *DishPopularityService) UpdatePopularityMetrics(dishID, dishSlug string) error {
	collection := s.Collection()
	now := time.Now()

	// Get counts from various collections
	favoriteService := NewUserFavoriteService()
	savedService := NewUserSavedDishService()
	interactionService := NewUserDishInteractionService()

	// Count favorites
	totalFavorites, _ := favoriteService.Collection().CountDocuments(context.TODO(), bson.M{
		"dishSlug": dishSlug,
		"deleted":  false,
	})

	// Count saved
	totalSaves, _ := savedService.Collection().CountDocuments(context.TODO(), bson.M{
		"dishSlug": dishSlug,
		"deleted":  false,
	})

	// Get interaction stats
	interactionFilter := bson.M{"dishSlug": dishSlug, "deleted": false}
	
	// Total views
	viewsPipeline := mongo.Pipeline{
		{{Key: "$match", Value: interactionFilter}},
		{{Key: "$group", Value: bson.M{
			"_id":        nil,
			"totalViews": bson.M{"$sum": "$viewCount"},
		}}},
	}
	viewsCursor, _ := interactionService.Collection().Aggregate(context.TODO(), viewsPipeline)
	var viewsResult []bson.M
	viewsCursor.All(context.TODO(), &viewsResult)
	var totalViews int64
	if len(viewsResult) > 0 {
		if val, ok := viewsResult[0]["totalViews"].(int32); ok {
			totalViews = int64(val)
		} else if val, ok := viewsResult[0]["totalViews"].(int64); ok {
			totalViews = val
		}
	}

	// Total cooked count
	cookedPipeline := mongo.Pipeline{
		{{Key: "$match", Value: interactionFilter}},
		{{Key: "$group", Value: bson.M{
			"_id":         nil,
			"totalCooked": bson.M{"$sum": "$cookedCount"},
		}}},
	}
	cookedCursor, _ := interactionService.Collection().Aggregate(context.TODO(), cookedPipeline)
	var cookedResult []bson.M
	cookedCursor.All(context.TODO(), &cookedResult)
	var totalCookedCount int64
	if len(cookedResult) > 0 {
		if val, ok := cookedResult[0]["totalCooked"].(int32); ok {
			totalCookedCount = int64(val)
		} else if val, ok := cookedResult[0]["totalCooked"].(int64); ok {
			totalCookedCount = val
		}
	}

	// Total shares
	sharesPipeline := mongo.Pipeline{
		{{Key: "$match", Value: interactionFilter}},
		{{Key: "$group", Value: bson.M{
			"_id":         nil,
			"totalShares": bson.M{"$sum": "$sharedCount"},
		}}},
	}
	sharesCursor, _ := interactionService.Collection().Aggregate(context.TODO(), sharesPipeline)
	var sharesResult []bson.M
	sharesCursor.All(context.TODO(), &sharesResult)
	var totalShares int64
	if len(sharesResult) > 0 {
		if val, ok := sharesResult[0]["totalShares"].(int32); ok {
			totalShares = int64(val)
		} else if val, ok := sharesResult[0]["totalShares"].(int64); ok {
			totalShares = val
		}
	}

	// Average rating
	ratingPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"dishSlug": dishSlug,
			"deleted":  false,
			"rating":   bson.M{"$exists": true, "$ne": nil},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":       nil,
			"avgRating": bson.M{"$avg": "$rating"},
			"count":     bson.M{"$sum": 1},
		}}},
	}
	ratingCursor, _ := interactionService.Collection().Aggregate(context.TODO(), ratingPipeline)
	var ratingResult []bson.M
	ratingCursor.All(context.TODO(), &ratingResult)
	var averageRating float64
	var totalRatings int64
	if len(ratingResult) > 0 {
		if val, ok := ratingResult[0]["avgRating"].(float64); ok {
			averageRating = val
		}
		if val, ok := ratingResult[0]["count"].(int32); ok {
			totalRatings = int64(val)
		} else if val, ok := ratingResult[0]["count"].(int64); ok {
			totalRatings = val
		}
	}

	// Calculate views for last 7 and 30 days (simplified - in real scenario would need timestamp filtering)
	viewsLast7Days := totalViews / 4   // Approximation
	viewsLast30Days := totalViews / 2  // Approximation

	// Calculate trending score (weighted recent activity)
	trendingScore := float64(viewsLast7Days)*2.0 + float64(totalFavorites)*1.5 + averageRating*10.0

	// Calculate overall popularity score
	popularityScore := float64(totalViews)*1.0 + 
		float64(totalFavorites)*5.0 + 
		float64(totalSaves)*3.0 + 
		float64(totalCookedCount)*7.0 + 
		float64(totalShares)*4.0 + 
		averageRating*20.0

	filter := bson.M{"dishSlug": dishSlug}
	update := bson.M{
		"$set": bson.M{
			"dishId":            dishID,
			"dishSlug":          dishSlug,
			"totalViews":        totalViews,
			"viewsLast7Days":    viewsLast7Days,
			"viewsLast30Days":   viewsLast30Days,
			"totalFavorites":    totalFavorites,
			"totalSaves":        totalSaves,
			"totalCookedCount":  totalCookedCount,
			"totalShares":       totalShares,
			"averageRating":     averageRating,
			"totalRatings":      totalRatings,
			"trendingScore":     trendingScore,
			"popularityScore":   popularityScore,
			"lastCalculatedAt":  now,
			"deleted":           false,
			"updatedAt":         now,
		},
		"$setOnInsert": bson.M{
			"createdAt": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	return err
}

func (s *DishPopularityService) GetTrendingDishes(dto model.TrendingDishesDto) ([]string, error) {
	collection := s.Collection()

	filter := bson.M{"deleted": false}
	
	sortField := "trendingScore"
	if dto.Period == "month" {
		sortField = "viewsLast30Days"
	} else if dto.Period == "week" {
		sortField = "viewsLast7Days"
	}

	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: -1}}).
		SetLimit(int64(dto.Limit)).
		SetProjection(bson.M{"dishSlug": 1})
	
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var popularities []model.DishPopularity
	if err = cursor.All(context.TODO(), &popularities); err != nil {
		return nil, err
	}

	slugs := make([]string, len(popularities))
	for i, pop := range popularities {
		slugs[i] = pop.DishSlug
	}

	return slugs, nil
}

func (s *DishPopularityService) GetPopularDishes(dto model.QueryDishPopularityDto) (*model.PaginationResponse, error) {
	collection := s.Collection()

	filter := bson.M{"deleted": false}
	
	if dto.MinAverageRating != nil {
		filter["averageRating"] = bson.M{"$gte": *dto.MinAverageRating}
	}
	
	if dto.MinTrendingScore != nil {
		filter["trendingScore"] = bson.M{"$gte": *dto.MinTrendingScore}
	}
	
	if dto.MinPopularityScore != nil {
		filter["popularityScore"] = bson.M{"$gte": *dto.MinPopularityScore}
	}

	sortField := "popularityScore"
	if dto.SortBy != nil {
		switch *dto.SortBy {
		case "trending":
			sortField = "trendingScore"
		case "rating":
			sortField = "averageRating"
		case "views":
			sortField = "totalViews"
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: sortField, Value: -1}})
	
	if dto.Limit > 0 {
		opts.SetLimit(int64(dto.Limit))
		opts.SetSkip(int64((dto.Page - 1) * dto.Limit))
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var popularities []model.DishPopularity
	if err = cursor.All(context.TODO(), &popularities); err != nil {
		return nil, err
	}

	totalItems, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	totalPages := float64(totalItems) / float64(dto.Limit)
	if totalItems%int64(dto.Limit) > 0 {
		totalPages = float64(int(totalPages) + 1)
	}

	return &model.PaginationResponse{
		Data: popularities,
		Metadata: model.CountMetaData{
			TotalItems:   totalItems,
			ItemCount:    len(popularities),
			ItemsPerPage: dto.Limit,
			TotalPages:   totalPages,
			CurrentPage:  dto.Page,
		},
	}, nil
}

func (s *DishPopularityService) GetPopularityMetrics(dishSlug string) (*model.DishPopularity, error) {
	collection := s.Collection()
	
	filter := bson.M{"dishSlug": dishSlug, "deleted": false}
	var popularity model.DishPopularity
	
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, result.Err()
	}

	if err := result.Decode(&popularity); err != nil {
		return nil, err
	}

	return &popularity, nil
}
