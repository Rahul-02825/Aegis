package database

import (
	"context"
	"os"
	"PrismX/logger"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client           *mongo.Client
	UserCollection   *mongo.Collection
	ConfigCollection *mongo.Collection
)

// var log = logger.InitLogger("app.log")

func ConnectDatabase() {

	// Load env
	if err := godotenv.Load(); err != nil {
		logger.Instance.Error("Error loading .env file "+ err.Error())
	}

	uri := os.Getenv("MONGO_URL")
	if uri == "" {
		logger.Instance.Warn("MONGO_URL not set")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	var err error
	client, err = mongo.Connect(context.Background(), opts)
	if err != nil {
		logger.Instance.Error("Mongo connect failed: "+ err.Error())
	}

	// Verify connection
	if err = client.Ping(context.Background(), nil); err != nil {
		logger.Instance.Error("Mongo ping failed: "+ err.Error())
	}

	db := client.Database("PrismX")
	UserCollection = db.Collection("User")
	ConfigCollection = db.Collection("Config")

	logger.Instance.Info("Successfully connected to MongoDB")
}


