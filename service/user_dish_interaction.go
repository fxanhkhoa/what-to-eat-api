package service

import (
	"context"
	"errors"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserDishInteractionService struct{}

func NewUserDishInteractionService() *UserDishInteractionService {
	return &UserDishInteractionService{}
}

func (s *UserDishInteractionService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.USER_DISH_INTERACTION_COLLECTION)
	return col
}

func (s *UserDishInteractionService) RecordView(dto model.RecordDishViewDto, profile *model.JwtCustomClaims) error {
	collection := s.Collection()
	now := time.Now()

	filter := bson.M{"userId": dto.UserID, "dishSlug": dto.DishSlug, "deleted": false}
	
	// Increment view count and update last viewed time
	update := bson.M{
		"$inc": bson.M{
			"viewCount": 1,
		},
		"$set": bson.M{
			"dishId":       dto.DishID,
			"lastViewedAt": now,
			"updatedAt":    now,
			"updatedBy":    profile.ID,
		},
		"$setOnInsert": bson.M{
			"userId":           dto.UserID,
			"dishSlug":         dto.DishSlug,
			"cooked":           false,
			"cookedCount":      0,
			"sharedCount":      0,
			"interactionScore": 0.0,
			"deleted":          false,
			"createdAt":        now,
			"createdBy":        profile.ID,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	// Recalculate interaction score
	return s.recalculateInteractionScore(dto.UserID, dto.DishSlug)
}

func (s *UserDishInteractionService) RecordCooked(dto model.RecordDishCookedDto, profile *model.JwtCustomClaims) error {
	collection := s.Collection()
	now := time.Now()

	filter := bson.M{"userId": dto.UserID, "dishSlug": dto.DishSlug, "deleted": false}
	
	update := bson.M{
		"$inc": bson.M{
			"cookedCount": 1,
		},
		"$set": bson.M{
			"dishId":       dto.DishID,
			"cooked":       true,
			"lastCookedAt": now,
			"updatedAt":    now,
			"updatedBy":    profile.ID,
		},
		"$setOnInsert": bson.M{
			"userId":           dto.UserID,
			"dishSlug":         dto.DishSlug,
			"viewCount":        0,
			"sharedCount":      0,
			"interactionScore": 0.0,
			"deleted":          false,
			"createdAt":        now,
			"createdBy":        profile.ID,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	return s.recalculateInteractionScore(dto.UserID, dto.DishSlug)
}

func (s *UserDishInteractionService) RateDish(dto model.RateDishDto, profile *model.JwtCustomClaims) error {
	collection := s.Collection()
	now := time.Now()

	if dto.Rating < 1 || dto.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	filter := bson.M{"userId": dto.UserID, "dishSlug": dto.DishSlug, "deleted": false}
	
	update := bson.M{
		"$set": bson.M{
			"dishId":    dto.DishID,
			"rating":    dto.Rating,
			"ratedAt":   now,
			"updatedAt": now,
			"updatedBy": profile.ID,
		},
		"$setOnInsert": bson.M{
			"userId":           dto.UserID,
			"dishSlug":         dto.DishSlug,
			"viewCount":        0,
			"cooked":           false,
			"cookedCount":      0,
			"sharedCount":      0,
			"interactionScore": 0.0,
			"deleted":          false,
			"createdAt":        now,
			"createdBy":        profile.ID,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	return s.recalculateInteractionScore(dto.UserID, dto.DishSlug)
}

func (s *UserDishInteractionService) RecordShare(dto model.RecordDishSharedDto, profile *model.JwtCustomClaims) error {
	collection := s.Collection()
	now := time.Now()

	filter := bson.M{"userId": dto.UserID, "dishSlug": dto.DishSlug, "deleted": false}
	
	update := bson.M{
		"$inc": bson.M{
			"sharedCount": 1,
		},
		"$set": bson.M{
			"dishId":       dto.DishID,
			"lastSharedAt": now,
			"updatedAt":    now,
			"updatedBy":    profile.ID,
		},
		"$setOnInsert": bson.M{
			"userId":           dto.UserID,
			"dishSlug":         dto.DishSlug,
			"viewCount":        0,
			"cooked":           false,
			"cookedCount":      0,
			"interactionScore": 0.0,
			"deleted":          false,
			"createdAt":        now,
			"createdBy":        profile.ID,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	return s.recalculateInteractionScore(dto.UserID, dto.DishSlug)
}

func (s *UserDishInteractionService) recalculateInteractionScore(userID, dishSlug string) error {
	collection := s.Collection()

	filter := bson.M{"userId": userID, "dishSlug": dishSlug, "deleted": false}
	var interaction model.UserDishInteraction
	
	if err := collection.FindOne(context.TODO(), filter).Decode(&interaction); err != nil {
		return err
	}

	// Calculate interaction score based on different actions with weights
	score := 0.0
	score += float64(interaction.ViewCount) * 1.0      // Each view = 1 point
	score += float64(interaction.CookedCount) * 10.0   // Each cooked = 10 points
	score += float64(interaction.SharedCount) * 5.0    // Each share = 5 points
	
	if interaction.Rating != nil {
		score += float64(*interaction.Rating) * 3.0     // Rating contributes (1-5) * 3 points
	}

	// Update the interaction score
	update := bson.M{
		"$set": bson.M{
			"interactionScore": score,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (s *UserDishInteractionService) GetUserInteractions(dto model.QueryUserDishInteractionDto) (*model.PaginationResponse, error) {
	collection := s.Collection()

	filter := bson.M{"deleted": false}
	
	if dto.UserID != nil {
		filter["userId"] = *dto.UserID
	}
	
	if dto.DishSlug != nil {
		filter["dishSlug"] = *dto.DishSlug
	}

	if dto.Cooked != nil {
		filter["cooked"] = *dto.Cooked
	}

	if dto.MinRating != nil {
		filter["rating"] = bson.M{"$gte": *dto.MinRating}
	}

	opts := options.Find().SetSort(bson.D{{Key: "interactionScore", Value: -1}})
	
	if dto.Limit > 0 {
		opts.SetLimit(int64(dto.Limit))
		opts.SetSkip(int64((dto.Page - 1) * dto.Limit))
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var interactions []model.UserDishInteraction
	if err = cursor.All(context.TODO(), &interactions); err != nil {
		return nil, err
	}

	totalItems, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := float64(totalItems) / float64(dto.Limit)
	if totalItems%int64(dto.Limit) > 0 {
		totalPages = float64(int(totalPages) + 1)
	}

	return &model.PaginationResponse{
		Data: interactions,
		Metadata: model.CountMetaData{
			TotalItems:   totalItems,
			ItemCount:    len(interactions),
			ItemsPerPage: dto.Limit,
			TotalPages:   totalPages,
			CurrentPage:  dto.Page,
		},
	}, nil
}

func (s *UserDishInteractionService) GetTopInteractedDishes(userID string, limit int) ([]string, error) {
	collection := s.Collection()
	
	filter := bson.M{"userId": userID, "deleted": false}
	opts := options.Find().
		SetSort(bson.D{{Key: "interactionScore", Value: -1}}).
		SetLimit(int64(limit)).
		SetProjection(bson.M{"dishSlug": 1})
	
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var interactions []model.UserDishInteraction
	if err = cursor.All(context.TODO(), &interactions); err != nil {
		return nil, err
	}

	slugs := make([]string, len(interactions))
	for i, interaction := range interactions {
		slugs[i] = interaction.DishSlug
	}

	return slugs, nil
}

func (s *UserDishInteractionService) GetAverageRating(userID string) (float64, error) {
	collection := s.Collection()
	
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"userId": userID, "deleted": false, "rating": bson.M{"$exists": true, "$ne": nil}}}},
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"avgRating": bson.M{"$avg": "$rating"},
		}}},
	}

	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(context.TODO())

	var result []bson.M
	if err = cursor.All(context.TODO(), &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	avgRating, ok := result[0]["avgRating"].(float64)
	if !ok {
		return 0, nil
	}

	return avgRating, nil
}
