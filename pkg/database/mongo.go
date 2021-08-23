package database

import (
	"context"
	"os"
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
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URL"))
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}
	d.Client = client

	err = d.Client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}

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

func (d *MongoDatabase) InsertCollection(data interface{}, name string) (string, error) {
	collection := d.Client.Database("strategies").Collection(name)

	response, err := collection.InsertOne(context.Background(), data, &options.InsertOneOptions{})
	if err != nil {
		return "", nil
	}

	return response.InsertedID.(string), nil
}

func (d *MongoDatabase) UpdateCollection(filter bson.M, data interface{}, name string) error {
	var dataMap bson.M
	dataBytes, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(dataBytes, &dataMap)
	if err != nil {
		return err
	}
	dataMapFull := bson.M{
		"$set": dataMap,
	}

	collection := d.Client.Database("strategies").Collection(name)
	_, err = collection.UpdateOne(context.Background(), filter, dataMapFull, &options.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (d *MongoDatabase) DeleteCollection(filter bson.M, name string) error {
	collection := d.Client.Database("strategies").Collection(name)
	_, err := collection.DeleteOne(context.Background(), filter, &options.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
