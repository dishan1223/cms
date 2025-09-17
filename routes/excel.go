package routes

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
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

	// Week order map
	weekOrder := map[string]int{
		"Saturday":  0,
		"Sunday":    1,
		"Monday":    2,
		"Tuesday":   3,
		"Wednesday": 4,
		"Thursday":  5,
		"Friday":    6,
	}

	// Study day mappings
	studyDayMap := map[string]string{
		"smw":     "Saturday, Monday, Wednesday",
		"stt":     "Saturday, Tuesday, Thursday",
		"regular": "Regular",
	}

	// Create Excel file
	f := excelize.NewFile()
	firstSheet := true

	for batch, batchStudents := range batchMap {
		// Map StudyDays codes to full names
		for i := range batchStudents {
			code := strings.ToLower(batchStudents[i].StudyDays)
			if full, ok := studyDayMap[code]; ok {
				batchStudents[i].StudyDays = full
			}
		}

		// Sort students by first study day, except "Regular" stays at the end
		sort.Slice(batchStudents, func(i, j int) bool {
			iDays := batchStudents[i].StudyDays
			jDays := batchStudents[j].StudyDays

			if iDays == "Regular" {
				return false
			}
			if jDays == "Regular" {
				return true
			}

			iFirst := strings.Split(iDays, ",")[0]
			jFirst := strings.Split(jDays, ",")[0]
			return weekOrder[iFirst] < weekOrder[jFirst]
		})

		sheetName := batch
		if firstSheet {
			f.SetSheetName(f.GetSheetName(0), sheetName)
			firstSheet = false
		} else {
			f.NewSheet(sheetName)
		}

		// Write headers
		headers := []string{"Name", "Phone Number", "Class", "Subject", "Payment Status", "Payment Amount", "Study Days"}
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
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), s.StudyDays)
		}
	}

	// Current month (e.g., September_2025)
	monthName := time.Now().Format("January_2006")

	// Add month to due_months where payment_status == false
	_, err = collection.UpdateMany(
		ctx,
		bson.M{"payment_status": false},
		bson.M{"$addToSet": bson.M{"due_months": monthName}},
	)
	if err != nil {
		log.Println("❌ Failed to add due months:", err)
	}

	// Reset payment_status for all students
	_, err = collection.UpdateMany(
		ctx,
		bson.M{},
		bson.M{"$set": bson.M{"payment_status": false}},
	)
	if err != nil {
		log.Println("❌ Failed to reset payments:", err)
	}

	// Dynamic filename
	filename := fmt.Sprintf("student_report_of_%s.xlsx", monthName)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Stream file
	if err := f.Write(c.Response().BodyWriter()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate Excel"})
	}

	return nil
}

