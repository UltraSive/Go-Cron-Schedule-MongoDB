package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Command struct for MongoDB document
type Command struct {
	Secret   string `bson:"secret"`
	Endpoint string `bson:"endpoint"`
	Schedule string `bson:"schedule"`
}

func main() {
	// MongoDB connection
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Error creating MongoDB client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Cron job scheduler
	c := cron.New()

	c.AddFunc("* * * * *", func() {
		// Fetch commands from MongoDB
		collection := client.Database("cron").Collection("commands")
		cur, err := collection.Find(ctx, nil)
		if err != nil {
			log.Printf("Error fetching commands from MongoDB: %v", err)
			return
		}
		defer cur.Close(ctx)

		for cur.Next(ctx) {
			var cmd Command
			err := cur.Decode(&cmd)
			if err != nil {
				log.Printf("Error decoding command from MongoDB: %v", err)
				continue
			}

			// Make GET request with authorization header
			req, err := http.NewRequest("GET", cmd.Endpoint, nil)
			if err != nil {
				log.Printf("Error creating HTTP request: %v", err)
				continue
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cmd.Secret))

			client := http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error making HTTP request: %v", err)
				continue
			}
			defer resp.Body.Close()

			log.Printf("Response from %s: %s", cmd.Endpoint, resp.Status)
		}
	})

	c.Start()

	// Keep the program running
	select {}
}
