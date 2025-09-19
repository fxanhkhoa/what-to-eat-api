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

type UserLoginTrackService struct{}

func (s *UserLoginTrackService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	return config.GetDBInstance().GetClient().Database(dbName).Collection(constants.USER_LOGIN_TRACKS_COLLECTION)
}

func (s *UserLoginTrackService) TrackLogin(userId, ip, userAgent string) error {
	collection := s.Collection()
	track := model.UserLoginTrack{
		UserID:    userId,
		LoginAt:   time.Now(),
		IP:        ip,
		UserAgent: userAgent,
	}
	_, err := collection.InsertOne(context.TODO(), track)
	return err
}

func (s *UserLoginTrackService) GetAllUserLogins(page, limit int64) ([]model.UserLoginTrack, int64, error) {
	collection := s.Collection()
	skip := (page - 1) * limit
	opts := options.Find()
	opts.SetLimit(limit)
	opts.SetSkip(skip)
	opts.SetSort(bson.D{{Key: "loginAt", Value: -1}})

	count, err := collection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	var tracks []model.UserLoginTrack
	err = cursor.All(context.TODO(), &tracks)
	return tracks, count, err
}

func (s *UserLoginTrackService) GetUserLogins(userId string, limit int64) ([]model.UserLoginTrack, int64, error) {
	collection := s.Collection()
	filter := bson.M{"userId": userId}
	opts := options.Find()
	opts.SetLimit(limit)
	opts.SetSort(bson.D{{Key: "loginAt", Value: -1}})

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var tracks []model.UserLoginTrack
	err = cursor.All(context.TODO(), &tracks)
	return tracks, count, err
}
