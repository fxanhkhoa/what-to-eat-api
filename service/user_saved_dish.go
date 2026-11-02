package service

import (
	"context"
	"errors"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserSavedDishService struct{}

func NewUserSavedDishService() *UserSavedDishService {
	return &UserSavedDishService{}
}

func (s *UserSavedDishService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.USER_SAVED_DISH_COLLECTION)
	return col
}

func (s *UserSavedDishService) SaveDish(dto model.CreateUserSavedDishDto, profile *model.JwtCustomClaims) (*model.UserSavedDish, error) {
	collection := s.Collection()
	now := time.Now()

	savedDish := model.UserSavedDish{
		UserID:    dto.UserID,
		DishID:    dto.DishID,
		DishSlug:  dto.DishSlug,
		Notes:     dto.Notes,
		Tags:      dto.Tags,
		Deleted:   false,
		CreatedAt: &now,
		CreatedBy: &profile.ID,
		UpdatedAt: &now,
		UpdatedBy: &profile.ID,
	}

	filter := bson.M{"userId": dto.UserID, "dishSlug": dto.DishSlug}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	
	update := bson.M{
		"$set": bson.M{
			"userId":    savedDish.UserID,
			"dishId":    savedDish.DishID,
			"dishSlug":  savedDish.DishSlug,
			"notes":     savedDish.Notes,
			"tags":      savedDish.Tags,
			"deleted":   false,
			"updatedAt": now,
			"updatedBy": profile.ID,
		},
		"$setOnInsert": bson.M{
			"createdAt": now,
			"createdBy": profile.ID,
		},
	}

	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil && result.Err() != mongo.ErrNoDocuments {
		return nil, result.Err()
	}

	decodeErr := result.Decode(&savedDish)
	if decodeErr != nil && decodeErr != mongo.ErrNoDocuments {
		return nil, decodeErr
	}

	return &savedDish, nil
}

func (s *UserSavedDishService) UpdateSavedDish(dto model.UpdateUserSavedDishDto, profile *model.JwtCustomClaims) (*model.UserSavedDish, error) {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(dto.ID)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "userId": dto.UserID, "deleted": false}
	update := bson.M{
		"$set": bson.M{
			"notes":     dto.Notes,
			"tags":      dto.Tags,
			"updatedAt": now,
			"updatedBy": profile.ID,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var savedDish model.UserSavedDish
	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}

	if err := result.Decode(&savedDish); err != nil {
		return nil, err
	}

	return &savedDish, nil
}

func (s *UserSavedDishService) RemoveSavedDish(dto model.DeleteUserSavedDishDto, profile *model.JwtCustomClaims) error {
	collection := s.Collection()
	now := time.Now()

	filter := bson.M{"userId": dto.UserID, "dishSlug": dto.DishSlug, "deleted": false}
	update := bson.M{
		"$set": bson.M{
			"deleted":   true,
			"deletedAt": now,
			"deletedBy": profile.ID,
		},
	}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("saved dish not found")
	}

	return nil
}

func (s *UserSavedDishService) GetUserSavedDishes(dto model.QueryUserSavedDishDto) (*model.PaginationResponse, error) {
	collection := s.Collection()

	filter := bson.M{"deleted": false}
	
	if dto.UserID != nil {
		filter["userId"] = *dto.UserID
	}
	
	if dto.DishSlug != nil {
		filter["dishSlug"] = *dto.DishSlug
	}

	if dto.Tags != nil && len(*dto.Tags) > 0 {
		filter["tags"] = bson.M{"$in": *dto.Tags}
	}

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	
	if dto.Limit > 0 {
		opts.SetLimit(int64(dto.Limit))
		opts.SetSkip(int64((dto.Page - 1) * dto.Limit))
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var savedDishes []model.UserSavedDish
	if err = cursor.All(context.TODO(), &savedDishes); err != nil {
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
		Data: savedDishes,
		Metadata: model.CountMetaData{
			TotalItems:   totalItems,
			ItemCount:    len(savedDishes),
			ItemsPerPage: dto.Limit,
			TotalPages:   totalPages,
			CurrentPage:  dto.Page,
		},
	}, nil
}

func (s *UserSavedDishService) IsSaved(userID, dishSlug string) (bool, error) {
	collection := s.Collection()
	
	filter := bson.M{"userId": userID, "dishSlug": dishSlug, "deleted": false}
	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *UserSavedDishService) GetUserSavedDishSlugs(userID string) ([]string, error) {
	collection := s.Collection()
	
	filter := bson.M{"userId": userID, "deleted": false}
	opts := options.Find().SetProjection(bson.M{"dishSlug": 1})
	
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var savedDishes []model.UserSavedDish
	if err = cursor.All(context.TODO(), &savedDishes); err != nil {
		return nil, err
	}

	slugs := make([]string, len(savedDishes))
	for i, saved := range savedDishes {
		slugs[i] = saved.DishSlug
	}

	return slugs, nil
}

func (s *UserSavedDishService) GetSavedCount(userID string) (int64, error) {
	collection := s.Collection()
	
	filter := bson.M{"userId": userID, "deleted": false}
	return collection.CountDocuments(context.TODO(), filter)
}
