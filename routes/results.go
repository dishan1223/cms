package routes

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

// StudentResult represents the incoming data from frontend
type StudentResult struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Class       string `json:"class"`
	BatchTime   string `json:"batch_time"`
	StudyDays   string `json:"study_days"`
	CQ          string `json:"cq"`
	MCQ         string `json:"mcq"`
}

// Handler function to receive results and return Excel
func SubmitResults(c *fiber.Ctx) error {
	var results []StudentResult

	// Parse request body
	if err := c.BodyParser(&results); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Run notifications
	for _, r := range results {
		notifyMarks(r)
	}

	// Convert marks to numbers for sorting
	type sortableResult struct {
		StudentResult
		Total int
	}

	var sortable []sortableResult
	for _, r := range results {
		cq, _ := strconv.Atoi(r.CQ)
		mcq, _ := strconv.Atoi(r.MCQ)
		total := cq + mcq
		sortable = append(sortable, sortableResult{StudentResult: r, Total: total})
	}

	// Sort by total marks descending
	sort.Slice(sortable, func(i, j int) bool {
		return sortable[i].Total > sortable[j].Total
	})

	// Create Excel file
	f := excelize.NewFile()
	sheet := "Results"
	f.NewSheet(sheet)

	// Header row
	headers := []string{"S.No", "Name", "Phone Number", "Class", "Batch Time", "Study Days", "CQ", "MCQ", "Total"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Data rows
	for i, r := range sortable {
		row := i + 2 // Excel rows start at 1
		values := []interface{}{
			i + 1,
			r.Name,
			r.PhoneNumber,
			r.Class,
			r.BatchTime,
			r.StudyDays,
			r.CQ,
			r.MCQ,
			r.Total,
		}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	// Stream the Excel file
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="results.xlsx"`)
	c.Set("File-Name", "results.xlsx")

	return f.Write(c.Response().BodyWriter())
}

// Example function to "notify" (you can later replace this with SMS/email)
func notifyMarks(r StudentResult) {
	fmt.Printf("📢 Student: %s (%s) | CQ: %s | MCQ: %s\n", r.Name, r.PhoneNumber, r.CQ, r.MCQ)
}

