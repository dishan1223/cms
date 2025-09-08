package main

import (
	"context"
	"log"
	"time"

	"github.com/dishan1223/cms/database"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	// Load ENV variables (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found, continuing with environment variables")
	}

	// Connect to MongoDB
	if err := database.ConnectDB(); err != nil {
		log.Fatal("❌ Database connection failed:", err)
	}

	// Run the payment reset task
	resetPayments()
}

func resetPayments() {
	collection := database.DB.Collection("students")
	filter := bson.M{}                                        // match all documents
	update := bson.M{"$set": bson.M{"payment_status": false}} // reset payment_status

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println("❌ Failed to reset payments:", err)
	} else {
		log.Println("✅ Payments reset to false for all students")
	}
}

