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

func MigrateBDT2UTC(collection *mongo.Collection, objectID primitive.ObjectID) {
	// Find the documents with the specific ObjectID
	cursor, err := collection.Find(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	// Iterate over the documents
	for cursor.Next(context.Background()) {
		var doc struct {
			ID                      primitive.ObjectID `bson:"_id"`
			StartDate               time.Time          `bson:"start_date"`
			EndDate                 time.Time          `bson:"end_date"`
			CreatedAt               time.Time          `bson:"created_at"`
			UpdatedAt               time.Time          `bson:"updated_at"`
			ClassStartDate          time.Time          `bson:"class_start_date"`
			ClassEndDate            time.Time          `bson:"class_end_date"`
			ContentAccessExpiryDate time.Time          `bson:"content_access_expiry_date"`
		}

		if err := cursor.Decode(&doc); err != nil {
			log.Fatal(err)
		}

		// Prepare updated fields, only including non-null and non-undefined values
		updatedFields := bson.M{}

		// Subtract 6 hours from each date
		adjustTime := func(t time.Time) time.Time {
			return t.Add(-6 * time.Hour)
		}

		if !doc.StartDate.IsZero() {
			updatedFields["start_date"] = adjustTime(doc.StartDate)
		}
		if !doc.EndDate.IsZero() {
			updatedFields["end_date"] = adjustTime(doc.EndDate)
		}
		if !doc.CreatedAt.IsZero() {
			updatedFields["created_at"] = adjustTime(doc.CreatedAt)
		}
		if !doc.UpdatedAt.IsZero() {
			updatedFields["updated_at"] = adjustTime(doc.UpdatedAt)
		}
		if !doc.ClassStartDate.IsZero() {
			updatedFields["class_start_date"] = adjustTime(doc.ClassStartDate)
		}
		if !doc.ClassEndDate.IsZero() {
			updatedFields["class_end_date"] = adjustTime(doc.ClassEndDate)
		}
		if !doc.ContentAccessExpiryDate.IsZero() {
			updatedFields["content_access_expiry_date"] = adjustTime(doc.ContentAccessExpiryDate)
		}

		// Only perform an update if there are fields to update
		if len(updatedFields) > 0 {
			// Perform the update
			_, err := collection.UpdateOne(
				context.Background(),
				bson.M{"_id": doc.ID},
				bson.M{"$set": updatedFields},
			)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Updated document with _id: %s\n", doc.ID.Hex())
		}
	}

	// Handle any cursor errors
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}
}
