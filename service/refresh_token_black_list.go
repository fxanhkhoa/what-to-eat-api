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
)

// RefreshTokenBlackListService provides CRUD operations for the refresh_token_black_list collection
type RefreshTokenBlackListService struct {
	Collection *mongo.Collection
}

func NewRefreshTokenBlackListService() *RefreshTokenBlackListService {
	dbName := config.GetDBInstance().GetDbName()
	return &RefreshTokenBlackListService{
		Collection: config.GetDBInstance().GetClient().Database(dbName).Collection(constants.REFRESH_TOKEN_BLACK_LIST_COLLECTION),
	}
}

func (s *RefreshTokenBlackListService) Create(token model.RefreshTokenBlackList) (primitive.ObjectID, error) {
	token.ID = primitive.NewObjectID()
	token.CreatedAt = time.Now()
	_, err := s.Collection.InsertOne(context.Background(), token)
	return token.ID, err
}

func (s *RefreshTokenBlackListService) GetByID(id primitive.ObjectID) (*model.RefreshTokenBlackList, error) {
	var result model.RefreshTokenBlackList
	err := s.Collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *RefreshTokenBlackListService) GetByToken(token string) (*model.RefreshTokenBlackList, error) {
	var result model.RefreshTokenBlackList
	err := s.Collection.FindOne(context.Background(), bson.M{"token": token}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *RefreshTokenBlackListService) Update(id primitive.ObjectID, update bson.M) error {
	res, err := s.Collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("not found")
	}
	return nil
}

func (s *RefreshTokenBlackListService) Delete(id primitive.ObjectID) error {
	res, err := s.Collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("not found")
	}
	return nil
}
