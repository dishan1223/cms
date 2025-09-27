package routes

import (
	"context"
    "time"

	"github.com/gofiber/fiber/v2"
	"github.com/dishan1223/cms/database"
	"github.com/dishan1223/cms/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetAllBatch(c *fiber.Ctx) error{
    batchCollection := database.DB.Collection("batches")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := batchCollection.Find(ctx, bson.M{})
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot fetch batches"})
    }
    defer cursor.Close(ctx)

    var batches []models.Batch
    if err = cursor.All(ctx, &batches); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot parse batches"})
    }

    return c.JSON(batches)
}


func AddBatch(c *fiber.Ctx) error {

    batchCollection := database.DB.Collection("batches")

    var batch models.Batch

    if err := c.BodyParser(&batch); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
    }

    batch.ID = primitive.NewObjectID()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
   
    _, err := batchCollection.InsertOne(ctx, batch)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot insert batch"})
    }


	return c.Status(fiber.StatusCreated).JSON(batch)
}


func DeleteBatch(c *fiber.Ctx) error {
    batchCollection := database.DB.Collection("batches")

    // Get the batch ID from the URL parameter
    idParam := c.Params("id")
    if idParam == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID is required"})
    }

    
    // Convert string ID to MongoDB ObjectID
    batchID, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
    }

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    

    // Delete the batch from the collection
    res, err := batchCollection.DeleteOne(ctx, bson.M{"_id": batchID})
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete batch"})
    }

    if res.DeletedCount == 0 {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Batch not found"})
    }

    return c.JSON(fiber.Map{"success": true, "message": "Batch deleted successfully"})
}

