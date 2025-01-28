package main

import (
	"github.com/Naheed-Rayhan/database_time_migration/Database"
	"github.com/Naheed-Rayhan/database_time_migration/Utils"
	"go.mongodb.org/mongo-driver/bson"
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

	fieldsToProcess := []string{}

	//--------------------------------------------------------------------------------------------
	// getting the collection
	coll1 := Database.GetCollection(client, "live-exam-dev", "model_tests")
	// List of fields to process and if maxUpdate is -1 then it will update all the documents
	fieldsToProcess = []string{"exam_date", "result_publish_time", "created_at", "updated_at"}
	Utils.MigrateBDT2UTC(coll1, 1, fieldsToProcess, bson.M{})

	coll2 := Database.GetCollection(client, "live-exam-dev", "model_test_session")
	// List of fields to process and if maxUpdate is -1 then it will update all the documents
	fieldsToProcess = []string{"start_time", "end_time", "created_at", "updated_at"}
	Utils.MigrateBDT2UTC(coll2, 1, fieldsToProcess, bson.M{})

	coll3 := Database.GetCollection(client, "live-exam-dev", "model_test_session_relation")
	// List of fields to process and if maxUpdate is -1 then it will update all the documents
	fieldsToProcess = []string{"created_at"}
	Utils.MigrateBDT2UTC(coll3, 1, fieldsToProcess, bson.M{})

	coll4 := Database.GetCollection(client, "live-exam-dev", "model_tests_result_state")
	// List of fields to process and if maxUpdate is -1 then it will update all the documents
	fieldsToProcess = []string{"created_at", "updated_at"}
	Utils.MigrateBDT2UTC(coll4, 1, fieldsToProcess, bson.M{})

	coll5 := Database.GetCollection(client, "academic-program-dev", "live_class")
	// List of fields to process and if maxUpdate is -1 then it will update all the documents
	fieldsToProcess = []string{"start_time", "end_time", "created_at", "updated_at"}
	Utils.MigrateBDT2UTC(coll5, 1, fieldsToProcess, bson.M{})

	coll6 := Database.GetCollection(client, "academic-program-dev", "lessons")
	// List of fields to process and if maxUpdate is -1 then it will update all the documents
	fieldsToProcess = []string{"start_time", "end_time", "created_at", "updated_at"}
	Utils.MigrateBDT2UTC(coll6, 1, fieldsToProcess, bson.M{"content_type": "HomeWork", "content_sub_type": "AnimatedVideo"})

	coll7 := Database.GetCollection(client, "academic-program-dev", "lessons")
	// List of fields to process and if maxUpdate is -1 then it will update all the documents
	fieldsToProcess = []string{"start_time", "end_time", "created_at", "updated_at"}
	Utils.MigrateBDT2UTC(coll7, 1, fieldsToProcess, bson.M{"content_type": "HomeWork", "content_sub_type": "Quiz"})

}
