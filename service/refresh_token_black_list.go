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
	col CollectionInterface
}

// NewRefreshTokenBlackListService creates a service with an optional injected collection.
// Pass nil to use the real MongoDB collection.
func NewRefreshTokenBlackListService(col CollectionInterface) *RefreshTokenBlackListService {
	return &RefreshTokenBlackListService{col: col}
}

func (s *RefreshTokenBlackListService) collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	return config.GetDBInstance().GetClient().Database(dbName).Collection(constants.REFRESH_TOKEN_BLACK_LIST_COLLECTION)
}

func (s *RefreshTokenBlackListService) getCol() CollectionInterface {
	if s.col != nil {
		return s.col
	}
	return NewMongoCollectionAdapter(s.collection())
}

func (s *RefreshTokenBlackListService) Create(token model.RefreshTokenBlackList) (primitive.ObjectID, error) {
	token.ID = primitive.NewObjectID()
	token.CreatedAt = time.Now()
	_, err := s.getCol().InsertOne(context.Background(), token)
	return token.ID, err
}

func (s *RefreshTokenBlackListService) GetByID(id primitive.ObjectID) (*model.RefreshTokenBlackList, error) {
	var result model.RefreshTokenBlackList
	err := s.getCol().FindOne(context.Background(), bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *RefreshTokenBlackListService) GetByToken(token string) (*model.RefreshTokenBlackList, error) {
	var result model.RefreshTokenBlackList
	err := s.getCol().FindOne(context.Background(), bson.M{"token": token}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *RefreshTokenBlackListService) Update(id primitive.ObjectID, update bson.M) error {
	res, err := s.getCol().UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("not found")
	}
	return nil
}

func (s *RefreshTokenBlackListService) Delete(id primitive.ObjectID) error {
	res, err := s.getCol().DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("not found")
	}
	return nil
}
