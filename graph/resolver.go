package graph

import (
	"context"
	"time"
	"what-to-eat/be/shared"

	"go.mongodb.org/mongo-driver/mongo"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{}

func Init(dbName string, collectionName string) (context.Context, *mongo.Collection) {
	ctxMongo, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	collection := shared.Client.Database(dbName).Collection(collectionName)
	return ctxMongo, collection
}
