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

type FoodShopDishService struct{}

func NewFoodShopDishService() *FoodShopDishService {
	return &FoodShopDishService{}
}

func (s *FoodShopDishService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	return config.GetDBInstance().GetClient().Database(dbName).Collection(constants.FOOD_SHOP_DISH_COLLECTION)
}

func (s *FoodShopDishService) Create(dto model.CreateFoodShopDishDto, profile *model.JwtCustomClaims) (*model.FoodShopDish, error) {
	collection := s.Collection()
	now := time.Now()

	entry := model.FoodShopDish{
		FoodShopID:  dto.FoodShopID,
		DishSlug:    dto.DishSlug,
		Price:       dto.Price,
		Currency:    dto.Currency,
		IsAvailable: dto.IsAvailable,
		Deleted:     false,
		CreatedAt:   &now,
		CreatedBy:   &profile.ID,
		UpdatedAt:   &now,
		UpdatedBy:   &profile.ID,
	}

	// Upsert: one record per shop+dish pair
	filter := bson.M{"foodShopId": dto.FoodShopID, "dishSlug": dto.DishSlug}
	update := bson.M{
		"$set": bson.M{
			"foodShopId":  entry.FoodShopID,
			"dishSlug":    entry.DishSlug,
			"price":       entry.Price,
			"currency":    entry.Currency,
			"isAvailable": entry.IsAvailable,
			"deleted":     false,
			"updatedAt":   now,
			"updatedBy":   profile.ID,
		},
		"$setOnInsert": bson.M{
			"createdAt": now,
			"createdBy": profile.ID,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	if err := result.Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *FoodShopDishService) Update(dto model.UpdateFoodShopDishDto, profile *model.JwtCustomClaims) (*model.FoodShopDish, error) {
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
	if dto.Price != nil {
		setFields["price"] = dto.Price
	}
	if dto.Currency != nil {
		setFields["currency"] = dto.Currency
	}
	if dto.IsAvailable != nil {
		setFields["isAvailable"] = dto.IsAvailable
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var entry model.FoodShopDish
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": setFields}, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	if err := result.Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *FoodShopDishService) Remove(id string, profile *model.JwtCustomClaims) (*model.FoodShopDish, error) {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	update := bson.M{
		"$set": bson.M{
			"deleted":   true,
			"deletedAt": now,
			"deletedBy": profile.ID,
		},
	}

	var entry model.FoodShopDish
	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	if err := result.Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *FoodShopDishService) FindOne(id string) (*model.FoodShopDish, error) {
	collection := s.Collection()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	var entry model.FoodShopDish
	err = collection.FindOne(context.TODO(), bson.M{"_id": objectID, "deleted": false}).Decode(&entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *FoodShopDishService) Find(query model.QueryFoodShopDishDto) ([]*model.FoodShopDish, int64, error) {
	collection := s.Collection()

	filter := bson.D{{Key: "deleted", Value: false}}

	if query.FoodShopID != nil && *query.FoodShopID != "" {
		filter = append(filter, bson.E{Key: "foodShopId", Value: *query.FoodShopID})
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

	var entries []*model.FoodShopDish
	if err = cursor.All(context.TODO(), &entries); err != nil {
		return nil, 0, err
	}

	return entries, count, nil
}
