package routes

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

// StudentResult represents the incoming data from frontend
type StudentResult struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	CQ          string `json:"cq"`
	MCQ         string `json:"mcq"`
	Total       int    // calculated
}

// Handler function to receive results
func SubmitResults(c *fiber.Ctx) error {
	var results []StudentResult

	// Parse request body
	if err := c.BodyParser(&results); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Calculate totals
	for i := range results {
		results[i].Total = parseMarks(results[i].CQ) + parseMarks(results[i].MCQ)
	}

	// Sort by total marks (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Total > results[j].Total
	})

	// Generate Excel
	if err := generateExcel(results); err != nil {
		log.Println("Excel generation error:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate Excel",
		})
	}

	// Notify each student
	for _, r := range results {
		notifyMarks(r)
	}

	return c.JSON(fiber.Map{
		"message": "Results received, Excel generated, and notifications triggered",
	})
}

// Helper to parse marks, treating "Absent" as 0
func parseMarks(m string) int {
	if m == "Absent" {
		return 0
	}
	val, err := strconv.Atoi(m)
	if err != nil {
		return 0
	}
	return val
}

// Generate Excel file with sorted results
func generateExcel(results []StudentResult) error {
	f := excelize.NewFile()
	sheet := "Results"
	index, _ := f.NewSheet(sheet)

	// Headers
	headers := []string{"Rank", "Name", "Phone Number", "CQ", "MCQ", "Total"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Fill rows
	for i, r := range results {
		row := i + 2 // row 2 onwards
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.Name)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.PhoneNumber)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), r.CQ)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), r.MCQ)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), r.Total)
	}

	f.SetActiveSheet(index)

	// Save file
	if err := f.SaveAs("results.xlsx"); err != nil {
		return err
	}
	return nil
}

// Example function to "notify" (you can later replace this with SMS/email)
func notifyMarks(r StudentResult) {
	fmt.Printf("📢 Student: %s (%s) | CQ: %s | MCQ: %s | Total: %d\n",
		r.Name, r.PhoneNumber, r.CQ, r.MCQ, r.Total)
}

