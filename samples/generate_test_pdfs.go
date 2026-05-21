//go:build ignore

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
)

func main() {
	forms := []func(string) error{
		createForm1Lines,
		createForm2Boxes,
		createForm3Underscores,
		createForm4Grids,
		createForm5NegativeBounds,
		createForm6Mixed,
		createForm7Multiline,
		createForm8TightLayout,
		createForm9DecorativeBanners,
		createForm10MultiPageKYC,
	}

	for i, createForm := range forms {
		filename := filepath.Join("pkg/pdfprocessor/testdata", fmt.Sprintf("test_form_%d.pdf", i+1))
		err := createForm(filename)
		if err != nil {
			log.Fatalf("Failed to generate %s: %v", filename, err)
		}
		log.Printf("Successfully generated test form: %s", filename)
	}
}

// Form 1: Horizontal Vector Lines Only
func createForm1Lines(filename string) error {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 1: Vector Lines Input Fields")

	pdf.SetFont("Arial", "", 12)
	pdf.SetLineWidth(1)

	pdf.Text(50, 100, "First Name:")
	pdf.Line(120, 100, 300, 100)

	pdf.Text(320, 100, "Last Name:")
	pdf.Line(390, 100, 550, 100)

	pdf.Text(50, 150, "Address Line 1:")
	pdf.Line(140, 150, 550, 150)

	pdf.Text(50, 200, "Phone:")
	pdf.Line(100, 200, 250, 200)

	pdf.Text(280, 200, "Email:")
	pdf.Line(320, 200, 550, 200)

	return pdf.OutputFileAndClose(filename)
}

// Form 2: Visual Rectangles Only
func createForm2Boxes(filename string) error {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 2: Rectangular Input Boxes")

	pdf.SetFont("Arial", "", 12)

	// Standard Comments Area
	pdf.Text(50, 100, "General Comments (Multiline):")
	pdf.Rect(50, 115, 500, 100, "D")

	// Message Body Area
	pdf.Text(50, 250, "Support Ticket Message Description:")
	pdf.Rect(50, 265, 500, 150, "D")

	// Small Note Area
	pdf.Text(50, 450, "Additional Notes:")
	pdf.Rect(50, 465, 500, 80, "D")

	return pdf.OutputFileAndClose(filename)
}

// Form 3: Text-based Underscores Only
func createForm3Underscores(filename string) error {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 3: Typography Underscores")

	pdf.SetFont("Arial", "", 12)
	pdf.Text(50, 100, "Username: _________________________")
	pdf.Text(50, 140, "Password: _________________________")
	pdf.Text(50, 180, "Recovery Question: _________________________________________________")
	pdf.Text(50, 220, "City: ___________________  State: _________  ZIP: ____________")

	return pdf.OutputFileAndClose(filename)
}

// Form 4: Row Grids and Small Table Boxes
func createForm4Grids(filename string) error {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 4: Row Grids and Input Tables")

	// Simple 3-column grid
	pdf.SetFont("Arial", "", 10)
	pdf.Text(50, 100, "Item Description")
	pdf.Text(300, 100, "Qty")
	pdf.Text(400, 100, "Price")

	// Draw visual comment table boundaries (large height cells)
	// Row 1
	pdf.Rect(50, 115, 230, 60, "D")
	pdf.Rect(280, 115, 100, 60, "D")
	pdf.Rect(380, 115, 150, 60, "D")

	// Row 2
	pdf.Rect(50, 185, 230, 60, "D")
	pdf.Rect(280, 185, 100, 60, "D")
	pdf.Rect(380, 185, 150, 60, "D")

	return pdf.OutputFileAndClose(filename)
}

// Form 5: Boxes Drawn with Negative Dimensions
func createForm5NegativeBounds(filename string) error {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 5: Negative Bounding Boxes")

	pdf.SetFont("Arial", "", 12)

	pdf.Text(50, 100, "Description A (drawn using negative width):")
	// pdf.Rect(x, y, w, h) -> we pass negative width
	pdf.Rect(550, 115, -500, 80, "D")

	pdf.Text(50, 250, "Description B (drawn using negative height):")
	// pdf.Rect(x, y, w, h) -> we pass negative height
	pdf.Rect(50, 345, 500, -80, "D")

	return pdf.OutputFileAndClose(filename)
}

// Form 6: Mixed Lines, Boxes, and Underscores
func createForm6Mixed(filename string) error {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 6: Mixed Input Format Document")

	pdf.SetFont("Arial", "", 11)
	pdf.SetLineWidth(1)

	// 1. Lines
	pdf.Text(50, 100, "Title: ")
	pdf.Line(80, 100, 200, 100)

	// 2. Underscores
	pdf.Text(220, 100, "Publisher: _____________________________")

	// 3. Large Description box
	pdf.Text(50, 150, "Abstract Description:")
	pdf.Rect(50, 165, 500, 120, "D")

	// 4. Additional Lines
	pdf.Text(50, 320, "Author A: ")
	pdf.Line(110, 320, 280, 320)

	pdf.Text(300, 320, "Author B: ")
	pdf.Line(360, 320, 530, 320)

	return pdf.OutputFileAndClose(filename)
}

// Form 7: Multiline Comment Blocks (Stacked elements)
func createForm7Multiline(filename string) error {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 7: Stacked Multiline Blocks")

	pdf.SetFont("Arial", "", 11)

	pdf.Text(50, 100, "Detailed Work Experience:")
	pdf.Rect(50, 115, 500, 100, "D")

	pdf.Text(50, 240, "Detailed Academic Credentials:")
	pdf.Rect(50, 255, 500, 100, "D")

	pdf.Text(50, 380, "Personal Statement:")
	pdf.Rect(50, 395, 500, 150, "D")

	return pdf.OutputFileAndClose(filename)
}

// Form 8: Tight Typography Spacing (Test metric accuracy)
func createForm8TightLayout(filename string) error {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 8: Tight Layout Fields")

	pdf.SetFont("Arial", "", 11)
	// Visual text sits extremely close to underscores
	pdf.Text(50, 100, "ID:___________")
	pdf.Text(150, 100, "Code:___________")
	pdf.Text(280, 100, "Ref:_______________")
	pdf.Text(420, 100, "Pin:_______")

	pdf.Text(50, 150, "Location:_____________________________________________")

	return pdf.OutputFileAndClose(filename)
}

// Form 9: Decorative Elements (To check filtering)
func createForm9DecorativeBanners(filename string) error {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 9: Decorative and Grid Filters")

	// 1. Thin grid lines (should be ignored)
	pdf.SetLineWidth(0.5)
	pdf.Line(50, 100, 550, 100)
	pdf.Line(50, 110, 550, 110)
	pdf.Line(50, 120, 550, 120)

	// 2. Visual title block header rectangle (thin height box - should be filtered)
	pdf.Rect(50, 150, 500, 20, "D")
	pdf.Text(60, 164, "SECTION 1: PERSONAL PROFILE (DO NOT WRITE HERE)")

	pdf.SetFont("Arial", "", 11)
	// 3. Actual input line (thick vector line - should be detected)
	pdf.SetLineWidth(1.5)
	pdf.Text(50, 220, "Full Legal Name:")
	pdf.Line(140, 220, 500, 220)

	// 4. Visual check box frames (tiny squares - should be filtered)
	pdf.Rect(50, 260, 15, 15, "D")
	pdf.Text(75, 272, "Option A")

	pdf.Rect(150, 260, 15, 15, "D")
	pdf.Text(175, 272, "Option B")

	return pdf.OutputFileAndClose(filename)
}

// Form 10: Multi-page Mixed KYC Document
func createForm10MultiPageKYC(filename string) error {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.SetLineWidth(1)
	pdf.SetFont("Arial", "", 11)

	// --- PAGE 1 ---
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 10: Multi-Page Mixed KYC (Page 1)")
	pdf.SetFont("Arial", "", 11)

	pdf.Text(50, 100, "First Name:")
	pdf.Line(120, 100, 280, 100)

	pdf.Text(300, 100, "Last Name:")
	pdf.Line(370, 100, 550, 100)

	pdf.Text(50, 150, "Citizenship: _____________________________")
	pdf.Text(50, 190, "Passport ID: _____________________________")

	// --- PAGE 2 ---
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 10: Multi-Page Mixed KYC (Page 2)")
	pdf.SetFont("Arial", "", 11)

	pdf.Text(50, 100, "Residential Address (Multiline):")
	pdf.Rect(50, 115, 500, 80, "D")

	pdf.Text(50, 220, "Mailing Address (if different):")
	pdf.Rect(50, 235, 500, 80, "D")

	// --- PAGE 3 ---
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(50, 50, "Form 10: Multi-Page Mixed KYC (Page 3)")
	pdf.SetFont("Arial", "", 11)

	pdf.Text(50, 100, "Income Declaration Comments:")
	pdf.Rect(50, 115, 500, 120, "D")

	pdf.Text(50, 280, "Signature Baseline:")
	pdf.Line(150, 280, 400, 280)

	pdf.Text(50, 340, "Date Signed: ___________________")

	return pdf.OutputFileAndClose(filename)
}
