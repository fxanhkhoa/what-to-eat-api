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

type UserDishLocationService struct{}

func NewUserDishLocationService() *UserDishLocationService {
	return &UserDishLocationService{}
}

func (s *UserDishLocationService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	return config.GetDBInstance().GetClient().Database(dbName).Collection(constants.USER_DISH_LOCATION_COLLECTION)
}

func (s *UserDishLocationService) Create(dto model.CreateUserDishLocationDto, profile *model.JwtCustomClaims) (*model.UserDishLocation, error) {
	collection := s.Collection()
	now := time.Now()

	location := model.UserDishLocation{
		UserID:     dto.UserID,
		DishSlug:   dto.DishSlug,
		FoodShopID: dto.FoodShopID,
		Name:       dto.Name,
		Address:    dto.Address,
		Latitude:   dto.Latitude,
		Longitude:  dto.Longitude,
		Notes:      dto.Notes,
		Deleted:    false,
		CreatedAt:  &now,
		CreatedBy:  &profile.ID,
		UpdatedAt:  &now,
		UpdatedBy:  &profile.ID,
	}

	result, err := collection.InsertOne(context.TODO(), location)
	if err != nil {
		return nil, err
	}

	location.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return &location, nil
}

func (s *UserDishLocationService) Update(dto model.UpdateUserDishLocationDto, profile *model.JwtCustomClaims) (*model.UserDishLocation, error) {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(dto.ID)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	setFields := bson.M{
		"updatedAt": now,
		"updatedBy": profile.ID,
	}
	if dto.Name != nil {
		setFields["name"] = dto.Name
	}
	if dto.Address != nil {
		setFields["address"] = dto.Address
	}
	if dto.Latitude != nil {
		setFields["latitude"] = dto.Latitude
	}
	if dto.Longitude != nil {
		setFields["longitude"] = dto.Longitude
	}
	if dto.Notes != nil {
		setFields["notes"] = dto.Notes
	}

	// Ownership check: only the owner can update their location
	filter := bson.M{"_id": objectID, "userId": profile.ID, "deleted": false}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var location model.UserDishLocation
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": setFields}, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	if err := result.Decode(&location); err != nil {
		return nil, err
	}
	return &location, nil
}

func (s *UserDishLocationService) Remove(id string, profile *model.JwtCustomClaims) (*model.UserDishLocation, error) {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	// Ownership check
	filter := bson.M{"_id": objectID, "userId": profile.ID, "deleted": false}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	update := bson.M{
		"$set": bson.M{
			"deleted":   true,
			"deletedAt": now,
			"deletedBy": profile.ID,
		},
	}

	var location model.UserDishLocation
	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	if err := result.Decode(&location); err != nil {
		return nil, err
	}
	return &location, nil
}

func (s *UserDishLocationService) FindOne(id string, userID string) (*model.UserDishLocation, error) {
	collection := s.Collection()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "userId": userID, "deleted": false}
	var location model.UserDishLocation
	err = collection.FindOne(context.TODO(), filter).Decode(&location)
	if err != nil {
		return nil, err
	}
	return &location, nil
}

func (s *UserDishLocationService) Find(query model.QueryUserDishLocationDto, userID string) ([]*model.UserDishLocation, int64, error) {
	collection := s.Collection()

	filter := bson.D{
		{Key: "userId", Value: userID},
		{Key: "deleted", Value: false},
	}

	if query.DishSlug != nil && *query.DishSlug != "" {
		filter = append(filter, bson.E{Key: "dishSlug", Value: *query.DishSlug})
	}

	page := query.BaseDto.Page
	if page < 1 {
		page = 1
	}
	limit := query.BaseDto.Limit
	if limit < 1 {
		limit = 10
	}
	skip := int64((page - 1) * limit)

	opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key: "createdAt", Value: -1}})

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())

	var locations []*model.UserDishLocation
	if err = cursor.All(context.TODO(), &locations); err != nil {
		return nil, 0, err
	}

	return locations, count, nil
}
