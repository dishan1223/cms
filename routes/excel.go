package routes

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dishan1223/cms/database"
	"github.com/dishan1223/cms/models"
	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func ExportStudents(c *fiber.Ctx) error {
	collection := database.DB.Collection("students")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch all students
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
	}
	defer cursor.Close(ctx)

	var students []models.Student
	if err := cursor.All(ctx, &students); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse students"})
	}

	// Group students by BatchTime
	batchMap := make(map[string][]models.Student)
	for _, s := range students {
		batchMap[s.BatchTime] = append(batchMap[s.BatchTime], s)
	}

	// Create Excel file
	f := excelize.NewFile()

	// Iterate over batches and create a sheet per batch
	firstSheet := true
	for batch, batchStudents := range batchMap {
		sheetName := batch
		if firstSheet {
			f.SetSheetName(f.GetSheetName(0), sheetName)
			firstSheet = false
		} else {
			f.NewSheet(sheetName)
		}

		// Write headers
		headers := []string{"Name", "Phone Number", "Class", "Subject", "Payment Status", "Payment Amount"}
		for i, h := range headers {
			col := string(rune('A' + i))
			f.SetCellValue(sheetName, col+"1", h)
		}

		// Write student rows
		for rowIdx, s := range batchStudents {
			row := rowIdx + 2
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), s.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), s.PhoneNumber)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), s.Class)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), s.Subject)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), s.PaymentStatus)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), s.PaymentAmount)
		}
	}

	// Reset all payment statuses to false after export
	_, err = collection.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"payment_status": false}})
	if err != nil {
		log.Println("❌ Failed to reset payments:", err)
	}

	// Dynamic filename with month and year
	monthName := time.Now().Format("January_2006")
	filename := fmt.Sprintf("student_report_of_%s.xlsx", monthName)

	// Set headers so browser downloads with the proper name
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Stream file
	if err := f.Write(c.Response().BodyWriter()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate Excel"})
	}

	return nil
}

