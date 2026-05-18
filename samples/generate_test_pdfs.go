package main

import (
	"log"
	"github.com/jung-kurt/gofpdf"
)

func main() {
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 50, "PDF Form Blanks Test - Free Open Source Version")

	pdf.SetFont("Arial", "", 12)
	pdf.Text(50, 100, "Name: ")
	
	// Case 1: Horizontal line
	pdf.SetLineWidth(1)
	pdf.Line(100, 100, 300, 100)

	pdf.Text(350, 100, "Date: ")
	
	// Case 2: Horizontal line
	pdf.Line(390, 100, 500, 100)

	pdf.Text(50, 150, "Comments:")

	// Case 3: Empty box
	pdf.Rect(50, 160, 400, 100, "D") // D = Draw boundary only

	// Case 4: Irrelevant vertical line
	pdf.SetLineWidth(2)
	pdf.Line(250, 300, 250, 400)

	// Case 5: Underscore chains (text-based blanks)
	pdf.Text(50, 430, "Username: _______________________")
	pdf.Text(50, 460, "Email: __________________________________")

	err := pdf.OutputFileAndClose("pkg/pdfprocessor/testdata/test_form.pdf")
	if err != nil {
		log.Fatalf("Fail writing file: %v", err)
	}
	log.Println("Created pkg/pdfprocessor/testdata/test_form.pdf")
}
