package main

import (
	"context"
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

	//
	fieldsToProcess := []string{}

	//-------------------------------------------MongoDB-------------------------------------------------

	// getting the collection
	coll1 := Database.GetCollection(client, "live-exam-dev", "model_tests")
	// List of fields to process and if maxUpdate is -1 then it will update all the documents
	fieldsToProcess = []string{"exam_date", "result_publish_time", "created_at", "updated_at"}
	Utils.MigrateBDT2UTC_mongoDB(coll1, 0, fieldsToProcess, bson.M{})

	//coll2 := Database.GetCollection(client, "live-exam-dev", "model_test_session")
	//// List of fields to process and if maxUpdate is -1 then it will update all the documents
	//fieldsToProcess = []string{"start_time", "end_time", "created_at", "updated_at"}
	//Utils.MigrateBDT2UTC_mongoDB(coll2, 1, fieldsToProcess, bson.M{})
	//
	//coll3 := Database.GetCollection(client, "live-exam-dev", "model_test_session_relation")
	//// List of fields to process and if maxUpdate is -1 then it will update all the documents
	//fieldsToProcess = []string{"created_at"}
	//Utils.MigrateBDT2UTC_mongoDB(coll3, 1, fieldsToProcess, bson.M{})
	//
	//coll4 := Database.GetCollection(client, "live-exam-dev", "model_tests_result_state")
	//// List of fields to process and if maxUpdate is -1 then it will update all the documents
	//fieldsToProcess = []string{"created_at", "updated_at"}
	//Utils.MigrateBDT2UTC_mongoDB(coll4, 1, fieldsToProcess, bson.M{})
	//
	//coll5 := Database.GetCollection(client, "academic-program-dev", "live_class")
	//// List of fields to process and if maxUpdate is -1 then it will update all the documents
	//fieldsToProcess = []string{"start_time", "end_time", "created_at", "updated_at"}
	//Utils.MigrateBDT2UTC_mongoDB(coll5, 1, fieldsToProcess, bson.M{})
	//
	//coll6 := Database.GetCollection(client, "academic-program-dev", "lessons")
	//// List of fields to process and if maxUpdate is -1 then it will update all the documents
	//fieldsToProcess = []string{"start_time", "end_time", "created_at", "updated_at"}
	//Utils.MigrateBDT2UTC_mongoDB(coll6, 1, fieldsToProcess, bson.M{"content_type": "HomeWork", "content_sub_type": "AnimatedVideo"})
	//
	//coll7 := Database.GetCollection(client, "academic-program-dev", "lessons")
	//// List of fields to process and if maxUpdate is -1 then it will update all the documents
	//fieldsToProcess = []string{"start_time", "end_time", "created_at", "updated_at"}
	//Utils.MigrateBDT2UTC_mongoDB(coll7, 1, fieldsToProcess, bson.M{"content_type": "HomeWork", "content_sub_type": "Quiz"})

	//-------------------------------------------ArangoDB-------------------------------------------------

	arangoDBendpointURL := os.Getenv("ARANGODB_URI")
	arangoDBusername := os.Getenv("ARANGODB_USERNAME")
	arangoDBpassword := os.Getenv("ARANGODB_PASSWORD")

	// Connecting to ArangoDB
	client2, err := Database.ConnectToArangoDB(context.Background(), arangoDBendpointURL, arangoDBusername, arangoDBpassword)
	if err != nil {
		log.Fatal(err)
	}

	// getting the database and collection
	db := Database.GetDB(client2, "shikho")
	fieldsToProcess = []string{"created_at", "otp_sent_time"} // Change the field name
	//Utils.MigrateBDT2UTC_arangoDB(context.Background(), fieldsToProcess, 3, db, "accounts")
	Utils.MigrateBDT2UTC_arangoDB2(context.Background(), fieldsToProcess, 1, db, "accounts")

}
