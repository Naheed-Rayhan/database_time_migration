package Utils

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver/v2/arangodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

func MigrateBDT2UTC_mongoDB(collection *mongo.Collection, maxUpdates int, fieldsToProcess []string, filter bson.M) {

	// Step 1: Fetch all documents that match the filter
	cursor, err := collection.Find(context.TODO(), filter)
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

func MigrateBDT2UTC_mongoDB_bulk(collection *mongo.Collection, maxUpdates int, fieldsToProcess []string, filter bson.M) {
	// Context with timeout for better resource management
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Step 1: Fetch documents in batches
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal("Failed to fetch documents:", err)
	}
	defer cursor.Close(ctx)

	// Bulk operation setup
	var bulkOps []mongo.WriteModel
	updateCount := 0

	// Process each document in the cursor
	for cursor.Next(ctx) {
		if updateCount >= maxUpdates && maxUpdates != -1 {
			fmt.Printf("Stopped after updating %d documents (maxUpdates limit reached).\n", maxUpdates)
			break
		}

		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Skipping document due to decoding error: %v\n", err)
			continue
		}

		updatePayload := bson.M{}
		hasUpdates := false

		// Process fields
		for _, field := range fieldsToProcess {
			if fieldValue, ok := doc[field]; ok {
				if dateTime, valid := fieldValue.(primitive.DateTime); valid {
					// Convert and adjust time
					newTime := dateTime.Time().Add(-6 * time.Hour)
					updatePayload[field] = newTime
					hasUpdates = true
				}
			}
		}

		// If no fields were updated, skip document
		if !hasUpdates {
			continue
		}

		// Prepare bulk update operation
		filter := bson.M{"_id": doc["_id"]}
		update := bson.M{"$set": updatePayload}
		bulkOps = append(bulkOps, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))

		// Execute batch update every 100 operations (adjustable)
		if len(bulkOps) >= 100 {
			res, err := collection.BulkWrite(ctx, bulkOps)
			if err != nil {
				log.Printf("Bulk update failed: %v\n", err)
			} else {
				updateCount += int(res.ModifiedCount)
			}
			bulkOps = nil // Reset batch
		}
	}

	// Process remaining bulk operations
	if len(bulkOps) > 0 {
		res, err := collection.BulkWrite(ctx, bulkOps)
		if err != nil {
			log.Printf("Final bulk update failed: %v\n", err)
		} else {
			updateCount += int(res.ModifiedCount)
		}
	}

	if err := cursor.Err(); err != nil {
		log.Fatal("Cursor error:", err)
	}

	log.Printf("Migration completed. Updated %d documents.\n", updateCount)
}

func MigrateBDT2UTC_arangoDB(ctx context.Context, fieldsToProcess []string, offset int, limit int, db arangodb.Database, collectionName string) {
	// Construct the AQL query for batch updates
	updateStatements := `"__tm": true,` // Add the __tm field

	for _, field := range fieldsToProcess {
		updateStatements += fmt.Sprintf(` "%s": HAS(doc, "%s") ? DATE_SUBTRACT(doc.%s, 6, "hour") : doc.%s,`, field, field, field, field)
	}
	updateStatements = updateStatements[:len(updateStatements)-1] // Remove last comma

	query := fmt.Sprintf(`
		FOR doc IN %s
		SORT doc._key ASC
		LIMIT %d, %d
		UPDATE doc WITH { %s } IN %s 
		RETURN { updated: NEW._key }
	`, collectionName, offset, limit, updateStatements, collectionName)

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer cursor.Close()

	// Track updates
	updateCount := 0
	for cursor.HasMore() {
		var result map[string]string
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			log.Printf("Failed to read update result: %v", err)
			continue
		}
		fmt.Printf("Updated document: %s\n", result["updated"])
		updateCount++
	}

	log.Printf("Total documents updated: %d\n", updateCount)
	log.Println("ArangoDB Migration completed.")
}
