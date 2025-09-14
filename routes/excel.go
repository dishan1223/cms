package routes

import (
	"context"
	"log"
	"time"
	"fmt"

	"github.com/dishan1223/cms/database"
	"github.com/dishan1223/cms/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/xuri/excelize/v2"
)

func ExportStudentsExcel(c *fiber.Ctx) error {
	collection := database.DB.Collection("students")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
	}
	defer cursor.Close(ctx)

	var students []models.Student
	if err := cursor.All(ctx, &students); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse students"})
	}

	// Create Excel file
	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetSheetName(f.GetSheetName(0), sheet)

	// Headers
	headers := []string{"Name", "Phone Number", "Batch Time", "Class", "Subject", "Payment Status", "Payment Amount"}
	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheet, col+"1", h)
	}

	// Group students by batch time
	row := 2
	for _, s := range students {
		f.SetCellValue(sheet, "A"+fmt.Sprint(row), s.Name)
		f.SetCellValue(sheet, "B"+fmt.Sprint(row), s.PhoneNumber)
		f.SetCellValue(sheet, "C"+fmt.Sprint(row), s.BatchTime)
		f.SetCellValue(sheet, "D"+fmt.Sprint(row), s.Class)
		f.SetCellValue(sheet, "E"+fmt.Sprint(row), s.Subject)
		f.SetCellValue(sheet, "F"+fmt.Sprint(row), s.PaymentStatus)
		f.SetCellValue(sheet, "G"+fmt.Sprint(row), s.PaymentAmount)
		row++
	}

	// Reset all payment statuses to false after export
	_, err = collection.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"payment_status": false}})
	if err != nil {
		log.Println("❌ Failed to reset payments:", err)
	}

	// Prepare dynamic filename
	monthName := time.Now().Format("January_2006") // e.g. September_2025
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

