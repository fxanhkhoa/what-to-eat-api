package shared

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DatabaseName string

func InitializeMongoDB() {
	connectionString := os.Getenv("MONGODB_CONNECTION_STRING")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(connectionString)
	Client, _ = mongo.Connect(ctx, clientOptions)
	DatabaseName = os.Getenv("DATABASE_NAME")

	indexUserCollection()
	indexRoleCollection()
	indexIngredientCollection()
	indexDishCollection()
}

func Init(collectionName string) (context.Context, *mongo.Collection) {
	ctxMongo, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	collection := Client.Database(DatabaseName).Collection(collectionName)
	return ctxMongo, collection
}

func indexUserCollection() {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: "text"}},
	}

	uniqueIDIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}, {Key: "googleID", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	collection := Client.Database(DatabaseName).Collection("Users")
	name, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		indexModel, uniqueIDIndexModel,
	})

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Created Index Users: %s \n", name)
}

func indexRoleCollection() {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	collection := Client.Database(DatabaseName).Collection("RolePermissions")
	name, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		indexModel,
	})

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Created Index RolePermissions: %s \n", name)
}

func indexIngredientCollection() {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "title.data", Value: "text"}},
		Options: options.Index().SetDefaultLanguage("en"),
	}

	uniqueSlug := mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	collection := Client.Database(DatabaseName).Collection("Ingredients")
	name, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		indexModel,
		uniqueSlug,
	})

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Created Index Ingredients: %s \n", name)
}

func indexDishCollection() {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "title.data", Value: "text"}},
		Options: options.Index().SetDefaultLanguage("en"),
	}

	uniqueSlug := mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	tagsIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "tags", Value: 1}},
	}

	collection := Client.Database(DatabaseName).Collection("Dishes")
	name, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		indexModel,
		uniqueSlug,
		tagsIndex,
	})

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Created Index Dishes: %s \n", name)
}
