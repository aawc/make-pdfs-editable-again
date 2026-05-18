package pdfprocessor

import (
	"os"
	"testing"
)

func TestDetectAndInject(t *testing.T) {
	inputFile := "testdata/test_form.pdf"
	outputFile := "testdata/test_form_editable.pdf"

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Skipf("File %s does not exist. Run 'go run samples/generate_test_pdfs.go' first", inputFile)
	}

	err := DetectAndInject(inputFile, outputFile)
	if err != nil {
		t.Fatalf("Failed to detect and inject: %v", err)
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Expected output file %s was not created", outputFile)
	}
	
	// We could also parse the written PDF to count fields, 
	// but testing if the process completes without error and creates a file is a great start.
}
