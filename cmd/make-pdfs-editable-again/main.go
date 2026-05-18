package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aawc/make-pdfs-editable-again/pkg/pdfprocessor"
)

func main() {
	var input string
	var output string

	flag.StringVar(&input, "input", "", "Path to the input PDF file")
	flag.StringVar(&output, "output", "", "Path for the output PDF file")
	flag.Parse()

	if input == "" {
		fmt.Fprintln(os.Stderr, "Error: --input is required.")
		flag.Usage()
		os.Exit(1)
	}

	if output == "" {
		output = input[:len(input)-len(".pdf")] + "_editable.pdf"
	}

	err := pdfprocessor.DetectAndInject(input, output)
	if err != nil {
		log.Fatalf("Fatal error processing PDF: %v", err)
	}

	fmt.Printf("Successfully saved editable PDF to: %s\n", output)
}
