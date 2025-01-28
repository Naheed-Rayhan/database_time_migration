package Database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func ConnectToMongoDB(uri string) (*mongo.Client, error) {
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	// ping the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to MongoDB")

	return client, nil
}

func GetCollection(client *mongo.Client, dbName, collectionName string) *mongo.Collection {
	// Get a handle for your collection
	collection := client.Database(dbName).Collection(collectionName)
	return collection
}

func DisconnectMongoDB(client *mongo.Client) error {
	// Disconnect from MongoDB
	err := client.Disconnect(context.TODO())
	if err != nil {
		return err
	}
	log.Printf("Disconnected from MongoDB")

	return nil
}
