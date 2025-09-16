package routes

import (
	"context"
    "time"
    "log"

	"github.com/gofiber/fiber/v2"
	"github.com/dishan1223/cms/database"
	"github.com/dishan1223/cms/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AddStudent handles POST requests to add a new student
func AddStudent(c *fiber.Ctx) error {
	// Get the collection here, after DB is initialized
	studentCollection := database.DB.Collection("students")

	student := new(models.Student)

	if err := c.BodyParser(student); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	student.ID = primitive.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := studentCollection.InsertOne(ctx, student)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot insert student"})
	}

	return c.Status(fiber.StatusCreated).JSON(student)
}

// get students by their unique id
func GetStudentByID(c *fiber.Ctx) error {
	id := c.Params("id")

	// Convert string ID → MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// Mongo collection
	collection := database.DB.Collection("students")

	// Find student
	var student models.Student
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&student)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	return c.JSON(student)
}


// GetStudents handles GET requests to fetch all students
func GetStudents(c *fiber.Ctx) error {
	studentCollection := database.DB.Collection("students")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := studentCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot fetch students"})
	}
	defer cursor.Close(ctx)

	var students []models.Student
	if err = cursor.All(ctx, &students); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot parse students"})
	}

	return c.JSON(students)
}

// DeleteStudent handles DELETE requests to remove a student by ID
func DeleteStudent(c *fiber.Ctx) error {
    studentCollection := database.DB.Collection("students")

    // Get the student ID from the URL parameter
    idParam := c.Params("id")
    if idParam == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID is required"})
    }

    // Convert string ID to MongoDB ObjectID
    studentID, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
    }

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Delete the student from the collection
    res, err := studentCollection.DeleteOne(ctx, bson.M{"_id": studentID})
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete student"})
    }

    if res.DeletedCount == 0 {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Student not found"})
    }

    // Return success message
    return c.JSON(fiber.Map{"message": "Student deleted successfully"})
}


// UpdateStudent handles PATCH requests to edit a student's info by ID
func UpdateStudent(c *fiber.Ctx) error {
	studentCollection := database.DB.Collection("students")

	// Get the student ID from the URL parameter
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID is required"})
	}

	// Convert string ID to MongoDB ObjectID
	studentID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// Parse request body JSON into a map
	updateData := make(map[string]interface{})
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare the update object
	update := bson.M{"$set": updateData}

	// Update the student in MongoDB
	res, err := studentCollection.UpdateByID(ctx, studentID, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update student"})
	}

	if res.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Student not found"})
	}

	return c.JSON(fiber.Map{"message": "Student updated successfully"})
}



func TogglePaymentStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	// Convert string ID to MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	collection := database.DB.Collection("students")

	// Find the student
	var student models.Student
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&student)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	currentMonth := time.Now().Format("January")

	// Toggle payment
	if student.PaymentStatus {
		// Paid → Unpaid
		student.PaymentStatus = false
	} else {
		// Unpaid → Paid
		student.PaymentStatus = true

		// Add current month to PaidMonths if not already present
		found := false
		for _, m := range student.PaidMonths {
			if m == currentMonth {
				found = true
				break
			}
		}
		if !found {
			student.PaidMonths = append(student.PaidMonths, currentMonth)
		}

		// Remove current month from DueMonths if it exists
		var updatedDue []string
		for _, m := range student.DueMonths {
			if m != currentMonth {
				updatedDue = append(updatedDue, m)
			}
		}
		student.DueMonths = updatedDue

		// Log the notification message to console
		notifyPayment(student)
	}

	// Update the student in MongoDB
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"payment_status": student.PaymentStatus,
				"paid_months":    student.PaidMonths,
				"due_months":     student.DueMonths,
			},
		},
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update student"})
	}

	return c.JSON(student)
}



// Just log the notification message like an SMS
func notifyPayment(student models.Student) {
    date :=  time.Now().Format("02-January-2006")
	message := "\n📢 Payment received for student: " +
		student.Name +
        "\n | Date: " + date +
        "\n | Class: " + student.Class +
		"\n | Subject: " + student.Subject +
		"\n | Batch: " + student.BatchTime +
		"\n | Phone: " + student.PhoneNumber

	log.Println(message)
}

// reset students' due months
func ResetDueMonths(c *fiber.Ctx) error {
	id := c.Params("id")

	// Convert string ID to MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	collection := database.DB.Collection("students")

	// Update the student's due_months field to null
	update := bson.M{
		"$set": bson.M{
			"due_months": nil,
		},
	}

	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": objID}, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to reset due months"})
	}

	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Due months reset successfully"})
}

