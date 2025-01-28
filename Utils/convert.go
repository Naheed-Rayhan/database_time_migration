package Utils

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

func MigrateBDT2UTC(collection *mongo.Collection, maxUpdates int, fieldsToProcess []string) {

	// Step 1: Fetch all documents
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal("Failed to fetch documents:", err)
	}
	defer cursor.Close(context.TODO())

	// Counter to track the number of updates
	updateCount := 0

	// Step 2: Iterate over each document
	for cursor.Next(context.TODO()) {
		// Stop if the maximum number of updates is reached (if maxUpdates is set to -1, it will update all documents)
		if updateCount >= maxUpdates && maxUpdates != -1 {
			fmt.Printf("Stopped after updating %d documents (maxUpdates limit reached).\n", maxUpdates)
			break
		}

		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Fatal("Failed to decode document:", err)
		}

		// Step 3: Prepare the update payload
		updatePayload := bson.M{}

		// Flag to track if any field was updated
		hasUpdates := false

		// Step 4: Process each field
		for _, field := range fieldsToProcess {
			// Check if the field exists in the document
			fieldValue, ok := result[field]
			if !ok {
				fmt.Printf("Skipping field %s in document %v: field is missing.\n", field, result["_id"])
				continue
			}

			// Ensure the field is a valid timestamp
			dateTime, ok := fieldValue.(primitive.DateTime)
			if !ok {
				fmt.Printf("Skipping field %s in document %v: field is not a valid timestamp.\n", field, result["_id"])
				continue
			}

			// Convert primitive.DateTime to time.Time
			currentTime := dateTime.Time()

			// Subtract 6 hours (adjust as needed for your timezone conversion)
			newTime := currentTime.Add(-6 * time.Hour)

			// Add the updated field to the payload
			updatePayload[field] = newTime
			hasUpdates = true
		}

		// Step 5: Skip if no fields were updated
		if !hasUpdates {
			fmt.Printf("No fields updated for document %v.\n", result["_id"])
			continue
		}

		// Step 6: Update the document with the new fields
		filter := bson.M{"_id": result["_id"]}
		update := bson.M{"$set": updatePayload}

		updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			log.Fatal("Failed to update document:", err)
		}

		if updateResult.MatchedCount == 0 {
			fmt.Println("No document matched the filter for document:", result["_id"])
		} else {
			fmt.Printf("Matched %v document(s) and updated %v document(s) for document: %v\n", updateResult.MatchedCount, updateResult.ModifiedCount, result["_id"])
			updateCount++
		}
	}

	if err := cursor.Err(); err != nil {
		log.Fatal("Cursor error:", err)
	}

	log.Printf("Migration completed. Updated %d documents.\n", updateCount)
}
