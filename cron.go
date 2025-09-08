package main

import (
	"context"
	"log"
	"time"

	"github.com/dishan1223/cms/database"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/joho/godotenv"
)

func main() {
	// Connect to MongoDB
	database.ConnectDB()

	// Load ENV
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ .env file not found, continuing with environment variables")
	}

	resetPayments()
}

func resetPayments() {
	collection := database.DB.Collection("students")
	filter := bson.M{}
	update := bson.M{"$set": bson.M{"payment_status": false}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println("❌ Failed to reset payments:", err)
	} else {
		log.Println("✅ Payments reset to false for all students")
	}
}

