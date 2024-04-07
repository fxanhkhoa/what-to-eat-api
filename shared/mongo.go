package shared

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DatabaseName string

func InitializeMongoDB() {
	uri := os.Getenv("MONGODB_CONNECTION_STRING")
	fmt.Println(uri)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	// defer func() {
	// 	if err = client.Disconnect(context.TODO()); err != nil {
	// 		panic(err)
	// 	}
	// }()

	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	Client = client
	dbName := os.Getenv("DATABASE_NAME")
	DatabaseName = dbName

	indexUserCollection()
	indexRoleCollection()
	indexIngredientCollection()
	indexDishCollection()
}

func Init(collectionName string) *mongo.Collection {
	collection := Client.Database(DatabaseName).Collection(collectionName)
	return collection
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
