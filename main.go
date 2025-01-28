package main

import (
	"github.com/Naheed-Rayhan/database_time_migration/Database"
	"github.com/Naheed-Rayhan/database_time_migration/Utils"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// getting the MongoDB URI from the environment variable
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable.")
	}

	// Connecting to MongoDB
	client, err := Database.ConnectToMongoDB(uri)
	if err != nil {
		log.Fatal(err)
	}

	// closing the Connection after scope
	defer func() {
		err := Database.DisconnectMongoDB(client)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// getting the collection
	coll := Database.GetCollection(client, "live-exam-dev", "model_tests")

	// List of fields to process and if maxUpdate is -1 then it will update all the documents
	fieldsToProcess := []string{"exam_date", "result_publish_time", "created_at", "updated_at"}
	Utils.MigrateBDT2UTC(coll, 1, fieldsToProcess)
}
