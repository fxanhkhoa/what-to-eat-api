package shared

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DatabaseName string

func InitializeMongoDB() {
	connectionString := os.Getenv("MONGODB_CONNECTION_STRING")
	fmt.Println(connectionString)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(connectionString)
	Client, _ = mongo.Connect(ctx, clientOptions)
	DatabaseName = os.Getenv("DATABASE_NAME")
}

func Init(dbName string, collectionName string) (context.Context, *mongo.Collection) {
	ctxMongo, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	collection := Client.Database(dbName).Collection(collectionName)
	return ctxMongo, collection
}
