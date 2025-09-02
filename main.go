package main


// TODO: Need to test cron and check if its reseting the data every month while in production server

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/robfig/cron/v3" // cron package
	"github.com/dishan1223/cms/database"
	"github.com/dishan1223/cms/routes"
	"go.mongodb.org/mongo-driver/bson"
	"context"
)


func resetPayments() {
	collection := database.DB.Collection("students")

	filter := bson.M{}
	update := bson.M{"$set": bson.M{"payment_status": false}} // match struct field

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println("❌ Failed to reset payments:", err)
	} else {
		log.Println("✅ Payments reset to false for all students")
	}
}


func main() {
	database.ConnectDB()

	app := fiber.New()

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Dishan! 🚀 GoFiber is running")
	})
	app.Get("/students", routes.GetStudents)
	app.Post("/students/new", routes.AddStudent)
	app.Delete("/students/delete/:id", routes.DeleteStudent)
	app.Patch("/students/edit/:id", routes.UpdateStudent)
	app.Patch("/students/payment/:id", routes.TogglePaymentStatus)

	// ---- Setup Cron Job ----
	c := cron.New()
	// Run at midnight on the 1st of every month
	_, err := c.AddFunc("0 0 1 * *", resetPayments)
	if err != nil {
		log.Fatal("❌ Error scheduling cron job:", err)
	}
	c.Start()

	log.Println("🚀 Server started on port 3000")
	app.Listen(":3000")
}

