# Make PDFs Editable Again

This CLI utility written in Go helps automatically detect visual "blanks" (like horizontal lines or empty boxed areas) in a static PDF file and overlays interactive digital form fields over them.

## Requirements

The project uses `github.com/pdfcpu/pdfcpu` and `github.com/jung-kurt/gofpdf`, robust open-source libraries under the Apache 2.0 and MIT licenses. No API keys or commercial licenses are required to run or compile this utility.

## Building the Tool

Ensure you have Go installed (version 1.18+ recommended).

```bash
go build -o bin/make-pdfs-editable-again cmd/make-pdfs-editable-again/main.go
```

## Usage

Run the compiled utility by providing an `--input` PDF.

```bash
mkdir -p out
./bin/make-pdfs-editable-again --input pkg/pdfprocessor/testdata/NRI_18.5.pdf --output out/my_editable_pdf.pdf
```

If `--output` is omitted, it defaults to the input filename with an `_editable.pdf` suffix.

## Running Tests and Samples

You can generate a test PDF with visual blanks to verify the detection logic:
```bash
go run samples/generate_test_pdfs.go
```
This generates `pkg/pdfprocessor/testdata/test_form.pdf`.

Then run the unit tests:
```bash
go test ./...
```

## How It Works

The utility parses the exact graphics state content streams inside the PDF rather than using OCR or imaging. It tracks the standard PDF drawing sequence:
1. It looks for `m` (moveTo) and `l` (lineTo) operators to trace flat uniform horizontal lines.
2. It looks for `re` (rectangle) operators for blank boxes suitable for multiline textual inputs.
3. It monitors text positioning (`Tm`, `Td`, `TD`) and text rendering (`Tj`, `TJ`) operators to isolate sequential underscore chains (`_______`) used as inline fill-in-the-blanks, dynamically calculating their precise visual offsets so that the injected form field overlays exactly on the underscores instead of the preceding text labels.
4. It filters visual items by appropriate dimensions (not too small, not an entire page background) and deduplicates adjacent overlapping entries.
5. For all matches, it appends a `Widget` annotation mapped to an `AcroForm` text entry.

## Limitations

Since PDFs can be composed of varying coordinate systems (CTM transformations), clipped paths, or image rasterizations, this pure content-stream parser relies on the document being constructed using standard vector path lines. Scanned documents without OCR lines or highly complex transformed forms might not yield perfect extraction boundaries.

## CI/CD & Security

The repository is secured and automated using GitHub Actions:
* **Security Analysis (Go Vulncheck)**: Every push and Pull Request is automatically scanned using Go's official `govulncheck` call-graph static analyzer to guarantee no dependencies with known vulnerabilities are introduced.
* **Automated Testing**: Unit tests and mock form generators run automatically on the CI pipeline to prevent regressions.
* **Continuous Delivery Release**: On every code commit pushed to the `main` branch, the release pipeline automatically generates a timestamped semantic tag in the format `vYYYY.MM.00N` (where the GitHub run number is zero-padded to 3 digits), compiles production-ready Go binaries for Windows (`.exe`), macOS, and Linux across AMD64 and ARM64 architectures, scans them using VirusTotal to verify they are safe from malware, and publishes a new GitHub Release attaching the binaries along with permanent public security analysis report links and the detailed commit changelog.
* **Static Web UI CD (GitHub Pages)**: Pushes to the `main` branch automatically compile our Go processor into WebAssembly, pre-compress the binary to `main.wasm.gz` using high-level Gzip compression (slashing the bundle size by 5x down to ~1.6MB), and deploy the responsive static dashboard hosted at `/web/` directly to **GitHub Pages**. The browser page utilizes native streaming decompression (`DecompressionStream`) to load the engine client-side with minimal network latency.


