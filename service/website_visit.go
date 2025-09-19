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

type WebsiteVisit struct {
	IP        string    `bson:"ip" json:"ip"`
	UserAgent string    `bson:"userAgent" json:"userAgent"`
	VisitedAt time.Time `bson:"visitedAt" json:"visitedAt"`
}

type WebsiteVisitService struct{}

func (s *WebsiteVisitService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	return config.GetDBInstance().GetClient().Database(dbName).Collection(constants.WEBSITE_VISIT_TRACKS_COLLECTION)
}

func (s *WebsiteVisitService) TrackVisit(ip, userAgent string) error {
	collection := s.Collection()
	visit := WebsiteVisit{
		IP:        ip,
		UserAgent: userAgent,
		VisitedAt: time.Now(),
	}
	_, err := collection.InsertOne(context.TODO(), visit)
	return err
}

func (s *WebsiteVisitService) CountVisits() (int64, error) {
	collection := s.Collection()
	return collection.CountDocuments(context.TODO(), bson.M{})
}

func (s *WebsiteVisitService) Find(query model.BaseDto) ([]*WebsiteVisit, int64, error) {
	collection := s.Collection()
	opts := options.Find().SetSort(bson.D{{Key: "visitedAt", Value: -1}}).SetSkip((int64(query.Page) - 1) * int64(query.Limit)).SetLimit(int64(query.Limit))
	filter := bson.D{}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())

	var visits []*WebsiteVisit
	if err = cursor.All(context.TODO(), &visits); err != nil {
		return nil, 0, err
	}

	return visits, count, nil
}
