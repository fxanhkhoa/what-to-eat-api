package service

import (
	"context"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
