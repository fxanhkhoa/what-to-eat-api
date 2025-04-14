package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client *mongo.Client
	DbName string
}

var singleInstance *MongoDB
var lock = &sync.Mutex{}

func GetDBInstance() *MongoDB {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			fmt.Println("Creating single instance now.")
			singleInstance = &MongoDB{}
			singleInstance.InitializeMongoDB()
		} else {
			fmt.Println("Single instance already created.")
		}
	}

	return singleInstance
}

func (d *MongoDB) InitializeMongoDB() {
	if err := godotenv.Load(); err != nil {
		fmt.Print(err)
	}

	uri := os.Getenv("MONGODB_CONNECTION_STRING")
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
	d.Client = client
	dbName := os.Getenv("DATABASE_NAME")
	d.DbName = dbName

	d.indexUserCollection()
	d.indexRoleCollection()
	d.indexIngredientCollection()
	d.indexDishCollection()
}

func (d *MongoDB) GetDbName() string {
	return d.DbName
}

func (d *MongoDB) GetClient() *mongo.Client {
	return d.Client
}

func (d *MongoDB) indexUserCollection() {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: "text"}},
	}

	uniqueIDIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}, {Key: "googleID", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	collection := d.Client.Database(d.DbName).Collection("Users")
	name, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		indexModel, uniqueIDIndexModel,
	})

	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Created Index Users: %s \n", name)
}

func (d *MongoDB) indexRoleCollection() {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	collection := d.Client.Database(d.DbName).Collection("RolePermissions")
	name, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		indexModel,
	})

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Created Index RolePermissions: %s \n", name)
}

func (d *MongoDB) indexIngredientCollection() {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "title.data", Value: "text"}},
		Options: options.Index().SetDefaultLanguage("en"),
	}

	uniqueSlug := mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	collection := d.Client.Database(d.DbName).Collection("Ingredients")
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

func (d *MongoDB) indexDishCollection() {
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

	collection := d.Client.Database(d.DbName).Collection("Dishes")
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
