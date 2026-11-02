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

type UserFavoriteService struct{}

func NewUserFavoriteService() *UserFavoriteService {
	return &UserFavoriteService{}
}

func (s *UserFavoriteService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.USER_FAVORITE_COLLECTION)
	return col
}

func (s *UserFavoriteService) AddFavorite(dto model.CreateUserFavoriteDto, profile *model.JwtCustomClaims) (*model.UserFavorite, error) {
	collection := s.Collection()
	now := time.Now()

	favorite := model.UserFavorite{
		UserID:    dto.UserID,
		DishID:    dto.DishID,
		DishSlug:  dto.DishSlug,
		Deleted:   false,
		CreatedAt: &now,
		CreatedBy: &profile.ID,
		UpdatedAt: &now,
		UpdatedBy: &profile.ID,
	}

	// Check if already exists (including soft-deleted)
	filter := bson.M{"userId": dto.UserID, "dishSlug": dto.DishSlug}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	
	update := bson.M{
		"$set": bson.M{
			"userId":    favorite.UserID,
			"dishId":    favorite.DishID,
			"dishSlug":  favorite.DishSlug,
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

	decodeErr := result.Decode(&favorite)
	if decodeErr != nil && decodeErr != mongo.ErrNoDocuments {
		return nil, decodeErr
	}

	return &favorite, nil
}

func (s *UserFavoriteService) RemoveFavorite(dto model.DeleteUserFavoriteDto, profile *model.JwtCustomClaims) error {
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
		return errors.New("favorite not found")
	}

	return nil
}

func (s *UserFavoriteService) GetUserFavorites(dto model.QueryUserFavoriteDto) (*model.PaginationResponse, error) {
	collection := s.Collection()

	filter := bson.M{"deleted": false}
	
	if dto.UserID != nil {
		filter["userId"] = *dto.UserID
	}
	
	if dto.DishSlug != nil {
		filter["dishSlug"] = *dto.DishSlug
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

	var favorites []model.UserFavorite
	if err = cursor.All(context.TODO(), &favorites); err != nil {
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
		Data: favorites,
		Metadata: model.CountMetaData{
			TotalItems:   totalItems,
			ItemCount:    len(favorites),
			ItemsPerPage: dto.Limit,
			TotalPages:   totalPages,
			CurrentPage:  dto.Page,
		},
	}, nil
}

func (s *UserFavoriteService) IsFavorite(userID, dishSlug string) (bool, error) {
	collection := s.Collection()
	
	filter := bson.M{"userId": userID, "dishSlug": dishSlug, "deleted": false}
	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *UserFavoriteService) GetUserFavoriteSlugs(userID string) ([]string, error) {
	collection := s.Collection()
	
	filter := bson.M{"userId": userID, "deleted": false}
	opts := options.Find().SetProjection(bson.M{"dishSlug": 1})
	
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var favorites []model.UserFavorite
	if err = cursor.All(context.TODO(), &favorites); err != nil {
		return nil, err
	}

	slugs := make([]string, len(favorites))
	for i, fav := range favorites {
		slugs[i] = fav.DishSlug
	}

	return slugs, nil
}

func (s *UserFavoriteService) GetFavoriteCount(userID string) (int64, error) {
	collection := s.Collection()
	
	filter := bson.M{"userId": userID, "deleted": false}
	return collection.CountDocuments(context.TODO(), filter)
}
