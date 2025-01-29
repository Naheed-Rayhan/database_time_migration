package Database

import (
	"context"
	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
	"log"
)

func ConnectToArangoDB(ctx context.Context, endpointURL string, username string, password string) (arangodb.Client, error) {

	endpoint := connection.NewRoundRobinEndpoints([]string{endpointURL})
	conn := connection.NewHttpConnection(connection.DefaultHTTPConfigurationWrapper(endpoint /*InsecureSkipVerify*/, false))

	// Add authentication
	auth := connection.NewBasicAuth(username, password)
	err := conn.SetAuthentication(auth)
	if err != nil {
		log.Fatalf("Failed to set authentication: %v", err)
	}

	// Create a client
	client := arangodb.NewClient(conn)
	log.Println("Connected to ArangoDB")

	return client, nil

}

func GetDBandCollection(client arangodb.Client, dbName string, collectionName string) (arangodb.Database, arangodb.Collection) {
	// Select database
	db, err := client.Database(nil, dbName) // Change database name
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	//Select collection
	collection, err := db.Collection(nil, collectionName) // Change collection name
	if err != nil {
		log.Fatalf("Failed to open collection: %v", err)
	}

	return db, collection
}
