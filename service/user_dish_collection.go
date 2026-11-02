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

type UserDishCollectionService struct{}

func NewUserDishCollectionService() *UserDishCollectionService {
	return &UserDishCollectionService{}
}

func (s *UserDishCollectionService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.USER_DISH_COLLECTION_COLLECTION)
	return col
}

func (s *UserDishCollectionService) Create(dto model.CreateUserDishCollectionDto, profile *model.JwtCustomClaims) (*model.UserDishCollection, error) {
	collection := s.Collection()
	now := time.Now()

	dishCollection := model.UserDishCollection{
		UserID:      dto.UserID,
		Name:        dto.Name,
		Description: dto.Description,
		Occasion:    dto.Occasion,
		EventDate:   dto.EventDate,
		DishSlugs:   dto.DishSlugs,
		Tags:        dto.Tags,
		IsPublic:    dto.IsPublic,
		Color:       dto.Color,
		Icon:        dto.Icon,
		SortOrder:   dto.SortOrder,
		Deleted:     false,
		CreatedAt:   &now,
		CreatedBy:   &profile.ID,
		UpdatedAt:   &now,
		UpdatedBy:   &profile.ID,
	}

	result, err := collection.InsertOne(context.TODO(), dishCollection)
	if err != nil {
		return nil, err
	}

	dishCollection.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return &dishCollection, nil
}

func (s *UserDishCollectionService) Update(dto model.UpdateUserDishCollectionDto, profile *model.JwtCustomClaims) (*model.UserDishCollection, error) {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(dto.ID)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "userId": dto.UserID, "deleted": false}
	update := bson.M{
		"$set": bson.M{
			"name":        dto.Name,
			"description": dto.Description,
			"occasion":    dto.Occasion,
			"eventDate":   dto.EventDate,
			"dishSlugs":   dto.DishSlugs,
			"tags":        dto.Tags,
			"isPublic":    dto.IsPublic,
			"color":       dto.Color,
			"icon":        dto.Icon,
			"sortOrder":   dto.SortOrder,
			"updatedAt":   now,
			"updatedBy":   profile.ID,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var dishCollection model.UserDishCollection
	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}

	if err := result.Decode(&dishCollection); err != nil {
		return nil, err
	}

	return &dishCollection, nil
}

func (s *UserDishCollectionService) Delete(collectionID, userID string, profile *model.JwtCustomClaims) error {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "userId": userID, "deleted": false}
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
		return errors.New("collection not found")
	}

	return nil
}

func (s *UserDishCollectionService) GetCollections(dto model.QueryUserDishCollectionDto) (*model.PaginationResponse, error) {
	collection := s.Collection()

	filter := bson.M{"deleted": false}
	
	if dto.UserID != nil {
		filter["userId"] = *dto.UserID
	}
	
	if dto.Occasion != nil {
		filter["occasion"] = *dto.Occasion
	}

	if dto.Tags != nil && len(*dto.Tags) > 0 {
		filter["tags"] = bson.M{"$in": *dto.Tags}
	}

	if dto.IsPublic != nil {
		filter["isPublic"] = *dto.IsPublic
	}

	if dto.Keyword != nil && *dto.Keyword != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": *dto.Keyword, "$options": "i"}},
			{"description": bson.M{"$regex": *dto.Keyword, "$options": "i"}},
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "sortOrder", Value: 1}, {Key: "createdAt", Value: -1}})
	
	if dto.Limit > 0 {
		opts.SetLimit(int64(dto.Limit))
		opts.SetSkip(int64((dto.Page - 1) * dto.Limit))
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var collections []model.UserDishCollection
	if err = cursor.All(context.TODO(), &collections); err != nil {
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
		Data: collections,
		Metadata: model.CountMetaData{
			TotalItems:   totalItems,
			ItemCount:    len(collections),
			ItemsPerPage: dto.Limit,
			TotalPages:   totalPages,
			CurrentPage:  dto.Page,
		},
	}, nil
}

func (s *UserDishCollectionService) GetCollectionByID(collectionID, userID string) (*model.UserDishCollection, error) {
	collection := s.Collection()

	objectID, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	// Only check userID if provided (allows public collections to be accessed by anyone)
	if userID != "" {
		filter["$or"] = []bson.M{
			{"userId": userID},
			{"isPublic": true},
		}
	} else {
		filter["isPublic"] = true
	}

	var dishCollection model.UserDishCollection
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}

	if err := result.Decode(&dishCollection); err != nil {
		return nil, err
	}

	return &dishCollection, nil
}

func (s *UserDishCollectionService) AddDish(dto model.AddDishToCollectionDto, profile *model.JwtCustomClaims) error {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(dto.CollectionID)
	if err != nil {
		return errors.New("invalid collection ID format")
	}

	filter := bson.M{"_id": objectID, "userId": dto.UserID, "deleted": false}
	update := bson.M{
		"$addToSet": bson.M{
			"dishSlugs": dto.DishSlug,
		},
		"$set": bson.M{
			"updatedAt": now,
			"updatedBy": profile.ID,
		},
	}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("collection not found")
	}

	return nil
}

func (s *UserDishCollectionService) RemoveDish(dto model.RemoveDishFromCollectionDto, profile *model.JwtCustomClaims) error {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(dto.CollectionID)
	if err != nil {
		return errors.New("invalid collection ID format")
	}

	filter := bson.M{"_id": objectID, "userId": dto.UserID, "deleted": false}
	update := bson.M{
		"$pull": bson.M{
			"dishSlugs": dto.DishSlug,
		},
		"$set": bson.M{
			"updatedAt": now,
			"updatedBy": profile.ID,
		},
	}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("collection not found")
	}

	return nil
}

func (s *UserDishCollectionService) ReorderDishes(dto model.ReorderDishesInCollectionDto, profile *model.JwtCustomClaims) error {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(dto.CollectionID)
	if err != nil {
		return errors.New("invalid collection ID format")
	}

	filter := bson.M{"_id": objectID, "userId": dto.UserID, "deleted": false}
	update := bson.M{
		"$set": bson.M{
			"dishSlugs": dto.DishSlugs,
			"updatedAt": now,
			"updatedBy": profile.ID,
		},
	}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("collection not found")
	}

	return nil
}

func (s *UserDishCollectionService) Duplicate(dto model.DuplicateCollectionDto, profile *model.JwtCustomClaims) (*model.UserDishCollection, error) {
	// Get the original collection
	originalCollection, err := s.GetCollectionByID(dto.CollectionID, dto.UserID)
	if err != nil {
		return nil, err
	}

	// Create new collection with copied data
	createDto := model.CreateUserDishCollectionDto{
		UserID:      dto.UserID,
		Name:        dto.NewName,
		Description: originalCollection.Description,
		Occasion:    originalCollection.Occasion,
		EventDate:   originalCollection.EventDate,
		DishSlugs:   originalCollection.DishSlugs,
		Tags:        originalCollection.Tags,
		IsPublic:    false, // Default to private
		Color:       originalCollection.Color,
		Icon:        originalCollection.Icon,
		SortOrder:   originalCollection.SortOrder,
	}

	if dto.CopyPublic {
		createDto.IsPublic = originalCollection.IsPublic
	}

	return s.Create(createDto, profile)
}

func (s *UserDishCollectionService) GetCollectionsBySlugs(userID string, dishSlugs []string) ([]model.UserDishCollection, error) {
	collection := s.Collection()

	filter := bson.M{
		"userId":    userID,
		"deleted":   false,
		"dishSlugs": bson.M{"$in": dishSlugs},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var collections []model.UserDishCollection
	if err = cursor.All(context.TODO(), &collections); err != nil {
		return nil, err
	}

	return collections, nil
}
