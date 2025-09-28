package main

import (
	"log"
	"os"

	"github.com/dishan1223/cms/database"
	"github.com/dishan1223/cms/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/dishan1223/cms/auth"
	"github.com/joho/godotenv"
)

func main() {
	// Load ENV variables (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found, continuing with environment variables")
	}

	// Connect to MongoDB
	if err := database.ConnectDB(); err != nil {
		log.Fatal("❌ Failed to connect to MongoDB:", err)
	}

	// Get PORT from env (Render provides $PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local testing
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

    // login route
    app.Get("/api/login", auth.LoginHandler)

    // students related routes
	app.Get("/students", routes.GetStudents)
    app.Get("/student/:id", routes.GetStudentByID)
	app.Post("/students/new", routes.AddStudent)
	app.Delete("/students/delete/:id", routes.DeleteStudent)
	app.Patch("/students/edit/:id", routes.UpdateStudent)
	app.Patch("/students/payment/:id", routes.TogglePaymentStatus)
    app.Get("/students/export", routes.ExportStudents)
    app.Patch("/students/reset-due-months/:id", routes.ResetDueMonths)

    // batch related routes
    app.Post("/api/batch/new", routes.AddBatch)
    app.Get("/api/batches", routes.GetAllBatch)
    app.Delete("/api/batch/:id", routes.DeleteBatch)


    app.Post("/api/submit-results", routes.SubmitResults)

	// Start server
	log.Println("🚀 Server starting on port " + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}

