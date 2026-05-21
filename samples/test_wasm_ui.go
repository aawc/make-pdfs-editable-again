package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// 1. Resolve paths
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	pdfPath := filepath.Join(pwd, "pkg/pdfprocessor/testdata/NRI_18.5.pdf")
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Fatalf("Test PDF does not exist: %s", pdfPath)
	}

	log.Println("Initializing headless Chrome automation via chromedp...")
	
	// 2. Configure allocator options (crucial for sandboxed containers)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.DisableGPU,
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithDebugf(log.Printf))
	defer cancel()

	// Add a strict timeout to prevent hanging tests
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var statusText string
	var buttonStyle string

	log.Println("Navigating to http://127.0.0.1:8080...")
	err = chromedp.Run(ctx,
		// Navigate
		chromedp.Navigate("http://127.0.0.1:8080"),
		// Wait for WASM initialization using explicit ByQuery CSS selector
		chromedp.WaitVisible("#status", chromedp.ByQuery),
		chromedp.Text("#status", &statusText, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal("Failed to load page: ", err)
	}

	log.Printf("Initial status: %q", statusText)
	
	// 3. Loop wait for WASM engine initialization if needed
	// (though usually it loads instantly on localhost)
	for i := 0; i < 10; i++ {
		if statusText == "WASM engine initialized. Ready to scan static forms." {
			break
		}
		time.Sleep(500 * time.Millisecond)
		err = chromedp.Run(ctx, chromedp.Text("#status", &statusText, chromedp.ByQuery))
		if err != nil {
			log.Fatal(err)
		}
	}

	if statusText != "WASM engine initialized. Ready to scan static forms." {
		log.Fatalf("WASM failed to initialize in time. Current status: %q", statusText)
	}

	log.Println("WASM engine ready! Uploading NRI_18.5.pdf...")
	
	// 4. Upload file and wait for success using explicit ByQuery selectors
	err = chromedp.Run(ctx,
		chromedp.SetUploadFiles(`#file-input`, []string{pdfPath}, chromedp.ByQuery),
		// Wait for processing to update the status
		chromedp.Sleep(2*time.Second), // Wait for Go WASM computation to finish
		chromedp.Text("#status", &statusText, chromedp.ByQuery),
		chromedp.AttributeValue("#download-btn", "style", &buttonStyle, nil, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal("Failed to upload or process PDF: ", err)
	}

	log.Printf("Status after upload: %q", statusText)
	log.Printf("Download button style: %q", buttonStyle)

	// 5. Final verification
	if statusText != "Successfully injected interactive form fields!" {
		log.Fatalf("Failed to inject fields successfully! Final status: %q", statusText)
	}

	log.Println("SUCCESS! In-browser WebAssembly scanned and generated interactive forms cleanly!")
}
