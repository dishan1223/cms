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

// Handler function to receive results and stream Excel file
func SubmitResults(c *fiber.Ctx) error {
	var results []StudentResult
	if err := c.BodyParser(&results); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Call notifyMarks for each student
	for _, s := range results {
		notifyMarks(s)
	}

	// Sort by total marks descending
	sort.Slice(results, func(i, j int) bool {
		totalI := parseMarks(results[i].CQ) + parseMarks(results[i].MCQ)
		totalJ := parseMarks(results[j].CQ) + parseMarks(results[j].MCQ)
		return totalI > totalJ
	})

	// Create Excel file
	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetSheetName(f.GetSheetName(0), sheet)
	f.SetCellValue(sheet, "A1", "Name")
	f.SetCellValue(sheet, "B1", "Phone")
	f.SetCellValue(sheet, "C1", "Class")
	f.SetCellValue(sheet, "D1", "Batch")
	f.SetCellValue(sheet, "E1", "Days")
	f.SetCellValue(sheet, "F1", "CQ")
	f.SetCellValue(sheet, "G1", "MCQ")
	f.SetCellValue(sheet, "H1", "Total")

	for i, s := range results {
		row := i + 2
		cq := parseMarks(s.CQ)
		mcq := parseMarks(s.MCQ)
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), s.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), s.PhoneNumber)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), s.Class)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), s.BatchTime)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), formatStudyDays(s.StudyDays))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), s.CQ)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), s.MCQ)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), cq+mcq)
	}

	// Stream Excel file to client
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=results.xlsx")
	buf, _ := f.WriteToBuffer()
	return c.SendStream(buf)
}

// notify function
func notifyMarks(r StudentResult) {
	fmt.Printf("📢 Student: %s (%s) | CQ: %s | MCQ: %s | Days: %s\n",
		r.Name, r.PhoneNumber, r.CQ, r.MCQ, formatStudyDays(r.StudyDays))
}

// Parse marks (handles "Absent")
func parseMarks(val string) int {
	if val == "Absent" {
		return 0
	}
	i, _ := strconv.Atoi(val)
	return i
}

// Convert SMW / STT to full names
func formatStudyDays(code string) string {
	switch code {
	case "SMW":
		return "Saturday, Monday, Wednesday"
	case "STT":
		return "Sunday, Tuesday, Thursday"
	default:
		return code
	}
}

