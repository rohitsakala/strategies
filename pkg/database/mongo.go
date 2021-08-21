package database

import (
	"context"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ Database = &MongoDatabase{}

type MongoDatabase struct {
	Client *mongo.Client
}

func (d *MongoDatabase) Connect() error {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}
	d.Client = client

	// Check the connection
	err = d.Client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}
	log.Printf("Connected to MongoDB")

	return nil
}

func (d *MongoDatabase) Disconnect() error {
	if err := d.Client.Disconnect(context.Background()); err != nil {
		return err
	}

	return nil
}

func (d *MongoDatabase) CreateCollection(name string) error {
	database := d.Client.Database("strategies")
	err := database.CreateCollection(context.Background(), name, &options.CreateCollectionOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "Collection already exists") {
			return err
		}
	}

	return nil
}

func (d *MongoDatabase) GetCollection(filter primitive.D, name string) (bson.M, error) {
	collection := d.Client.Database("strategies").Collection(name)

	singleResult := collection.FindOne(context.Background(), filter, &options.FindOneOptions{})
	var resultDoc bson.M
	err := singleResult.Decode(&resultDoc)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
	}

	return resultDoc, nil
}

func (d *MongoDatabase) InsertCollection(data interface{}, name string) error {
	collection := d.Client.Database("strategies").Collection(name)

	_, err := collection.InsertOne(context.Background(), data, &options.InsertOneOptions{})
	if err != nil {
		return nil
	}

	return nil
}

func (d *MongoDatabase) UpdateCollection(data interface{}, name string) error {
	collection := d.Client.Database("strategies").Collection(name)

	_, err := collection.UpdateOne(context.Background(), data, &options.InsertOneOptions{})
	if err != nil {
		return nil
	}

	return nil
}
