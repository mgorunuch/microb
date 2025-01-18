package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Document struct {
	ID             primitive.ObjectID `bson:"_id"`
	Data           Data               `bson:"data"`
	IsValid        bool               `bson:"is_valid"`
	ScoreID        string             `bson:"score_id"`
	Timestamp      time.Time          `bson:"timestamp"`
	StatusResponse *string            `bson:"status_response,omitempty"`
}

type Data struct {
	Status bool `bson:"status"`
}

type StatusRequest struct {
	ScoreID string `json:"score_id"`
}

func main() {
	ctx := context.Background()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to MongoDB: %v", err))
	}
	defer client.Disconnect(ctx)

	collection := client.Database("yasss_om-api_com").Collection("api_cache")

	// Find all documents where data.status is true
	filter := bson.M{
		"data.status": true,
	}

	fmt.Printf("Starting document processing...\n")

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		panic(fmt.Sprintf("Failed to query documents: %v", err))
	}
	defer cursor.Close(ctx)

	processedCount := 0
	successCount := 0

	// Process each document
	for cursor.Next(ctx) {
		var doc Document
		if err := cursor.Decode(&doc); err != nil {
			fmt.Printf("Failed to decode document: %v\n", err)
			continue
		}

		processedCount++
		fmt.Printf("\nProcessing document %d - ID: %s, ScoreID: %s\n",
			processedCount, doc.ID.Hex(), doc.ScoreID)

		// Make API request
		status, err := fetchStatus(doc.ScoreID)
		if err != nil {
			fmt.Printf("Failed to fetch status for score_id %s: %v\n", doc.ScoreID, err)
			continue
		}

		// Update document with status response
		updateFilter := bson.M{"_id": doc.ID}
		update := bson.M{
			"$set": bson.M{
				"status_response": status,
			},
		}

		opts := options.Update().SetUpsert(false)
		result, err := collection.UpdateOne(ctx, updateFilter, update, opts)
		if err != nil {
			fmt.Printf("Failed to update document %s: %v\n", doc.ID.Hex(), err)
			continue
		}

		// Verify the update
		var updatedDoc Document
		err = collection.FindOne(ctx, updateFilter).Decode(&updatedDoc)
		if err != nil {
			fmt.Printf("Failed to verify update for document %s: %v\n", doc.ID.Hex(), err)
		} else {
			if updatedDoc.StatusResponse != nil {
				successCount++
				fmt.Printf("Successfully updated document %s with status response\n", doc.ID.Hex())
			}
		}

		fmt.Printf("Update stats - Modified: %d, Matched: %d\n",
			result.ModifiedCount, result.MatchedCount)
	}

	fmt.Printf("\nProcessing complete!\n")
	fmt.Printf("Total documents processed: %d\n", processedCount)
	fmt.Printf("Successfully updated: %d\n", successCount)

	if err := cursor.Err(); err != nil {
		panic(fmt.Sprintf("Cursor error: %v", err))
	}
}

func fetchStatus(scoreID string) (string, error) {
	url := "https://yasss.om-api.com/score/advanced/status"

	reqBody := StatusRequest{
		ScoreID: scoreID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}
