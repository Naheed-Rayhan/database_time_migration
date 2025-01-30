package Utils

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver/v2/arangodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strings"
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

func MigrateBDT2UTC_arangoDB(ctx context.Context, fieldsToProcess []string, maxUpdates int, db arangodb.Database, collectionName string) {

	//Select collection
	collection, err := db.Collection(nil, collectionName) // Change collection name
	if err != nil {
		log.Fatalf("Failed to open collection: %v", err)
	}

	filterConditions := ""
	for i, field := range fieldsToProcess {
		if i > 0 {
			filterConditions += " AND "
		}
		filterConditions += fmt.Sprintf("doc.%s != null", field)
	}

	//query := fmt.Sprintf("FOR doc IN %s FILTER %s AND doc.id == \"%s\" RETURN doc", collectionName, filterConditions, "005X04AV19")
	query := fmt.Sprintf("FOR doc IN %s FILTER %s RETURN doc", collectionName, filterConditions)
	fmt.Println(query)

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer cursor.Close()

	updateCount := 0

	// Iterate over the cursor to process each document
	for cursor.HasMore() {

		// Stop if we've reached the maximum number of updates (if maxUpdates is set to -1, it will update all documents)
		if updateCount >= maxUpdates && maxUpdates != -1 {
			log.Println("Stopped after updating", maxUpdates, "documents (maxUpdates limit reached).")
			break
		}

		//imporvement
		var doc map[string]interface{}
		meta, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			log.Printf("Failed to read document: %v", err)
			continue
		}

		// Process each field to convert BDT to UTC
		for _, field := range fieldsToProcess {
			fieldTime, err := time.Parse(time.RFC3339, doc[field].(string))
			if err != nil {
				log.Printf("Error parsing time for document %s: %v", meta.Key, err)
				continue
			}
			// Subtract 6 hours to convert BDT to UTC
			doc[field] = fieldTime.Add(-6 * time.Hour).Format(time.RFC3339)
		}

		// Update the document in the collection
		_, err = collection.UpdateDocument(ctx, meta.Key, doc)
		if err != nil {
			log.Printf("Failed to update document %s: %v", meta.Key, err)
		} else {
			fmt.Printf("Updated document %s with new times\n", meta.Key)
			updateCount++
		}

	}

	log.Printf("Total documents updated: %d\n", updateCount)
	log.Println("arangoDB Migration completed.")

}

func MigrateBDT2UTC_arangoDB2(ctx context.Context, fieldsToProcess []string, maxUpdates int, db arangodb.Database, collectionName string) {
	//collection, err := db.Collection(nil, collectionName)
	//if err != nil {
	//	log.Fatalf("Failed to open collection: %v", err)
	//}

	// Construct the AQL query for batch updates
	updateStatements := ""
	for _, field := range fieldsToProcess {
		updateStatements += fmt.Sprintf(` "%s": DATE_SUBTRACT(doc.%s, 6, "hour"),`, field, field)
	}
	updateStatements = updateStatements[:len(updateStatements)-1] // Remove last comma

	query := fmt.Sprintf(`
		FOR doc IN %s 
		FILTER %s
		
		LIMIT %d
		UPDATE doc WITH { %s } IN %s 
		RETURN { updated: NEW._key }
	`, collectionName, generateFilter(fieldsToProcess), maxUpdates, updateStatements, collectionName)

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

// Helper function to generate dynamic filter conditions
func generateFilter(fields []string) string {
	conditions := []string{}
	for _, field := range fields {
		conditions = append(conditions, fmt.Sprintf(`doc.%s != null`, field))
	}
	return strings.Join(conditions, " AND ")
}
