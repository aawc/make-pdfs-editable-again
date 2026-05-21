package pdfprocessor

import (
	"fmt"
	"os"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/form"
)

func TestDetectAndInjectComprehensiveSuite(t *testing.T) {
	// Loop and scan all 11 distinct test form templates
	for i := 1; i <= 11; i++ {
		inputFile := fmt.Sprintf("testdata/test_form_%d.pdf", i)
		outputFile := fmt.Sprintf("testdata/test_form_%d_editable.pdf", i)

		t.Run(fmt.Sprintf("Form_%d", i), func(t *testing.T) {
			if _, err := os.Stat(inputFile); os.IsNotExist(err) {
				t.Skipf("File %s does not exist. Run 'go run samples/generate_test_pdfs.go' first", inputFile)
			}

			// 1. Execute the field detection and in-memory overlay injection
			err := DetectAndInject(inputFile, outputFile)
			if err != nil {
				t.Fatalf("Failed to detect and inject fields for %s: %v", inputFile, err)
			}

			// 2. Verify the output file is successfully written
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Fatalf("Expected output file %s was not created", outputFile)
			}

			// 3. Retrieve context and validate structural PDF AcroForm integrity
			ctx, err := api.ReadContextFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read processed PDF context for validation: %v", err)
			}

			err = api.ValidateContext(ctx)
			if err != nil {
				t.Fatalf("Processed PDF failed structural validation checks: %v", err)
			}

			// 4. Confirm AcroForm interactive fields were successfully injected
			fields, err := form.ListFormFields(ctx)
			if err != nil {
				t.Fatalf("Failed to list AcroForm interactive fields: %v", err)
			}

			t.Logf("Successfully scanned and injected %d interactive fields into Form %d!", len(fields), i)
			if len(fields) == 0 {
				t.Fatalf("Warning: Injected 0 interactive fields inside Form %d", i)
			}
		})
	}
}
