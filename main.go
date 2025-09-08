package main

import (
	"log"
	"os"

	"github.com/dishan1223/cms/database"
	"github.com/dishan1223/cms/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Connect to MongoDB
	database.ConnectDB()

	// Load ENV variables
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ .env file not found, continuing with environment variables")
	}

	// Get PORT from env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := fiber.New()

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PATCH, DELETE",
		AllowHeaders: "Content-Type, Authorization, Accept, Origin",
	}))

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Dishan! 🚀 GoFiber is running")
	})
	app.Get("/students", routes.GetStudents)
	app.Post("/students/new", routes.AddStudent)
	app.Delete("/students/delete/:id", routes.DeleteStudent)
	app.Patch("/students/edit/:id", routes.UpdateStudent)
	app.Patch("/students/payment/:id", routes.TogglePaymentStatus)

	// Start server
	log.Println("🚀 Server started on port " + port)
	err = app.Listen(":" + port)
	if err != nil {
		log.Fatal(err)
	}
}

