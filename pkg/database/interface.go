package database

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Database interface {
	Connect() error
	Disconnect() error

	// Collections
	CreateCollection(name string) error
	GetCollection(filter primitive.D, name string) (bson.M, error)
	InsertCollection(data interface{}, name string) error
	UpdateCollection(data interface{}, name string) error
}
