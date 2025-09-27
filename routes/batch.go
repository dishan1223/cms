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

