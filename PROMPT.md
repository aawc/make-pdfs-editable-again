# Recreation Prompt

You can use the following prompt with any advanced LLM (like GPT-4, Claude 3, or Gemini) to recreate this exact tool from scratch.

---

**Prompt:**

> Write a cross-platform command-line utility in Go that takes an existing PDF document as input and automatically adds interactive digital text form fields wherever it detects visual "blanks" (such as horizontal lines, empty rectangles, or consecutive text underscore chains). 
> 
> **Requirements**:
> 1. Do not use PyPI or NPM packages. Keep it strictly Go (this allows distribution as a single binary).
> 2. Use `github.com/pdfcpu/pdfcpu` to parse the PDF geometry, inject form fields, and output the PDF.
> 3. Create a logic module (e.g., `pkg/pdfprocessor/processor.go`) that parses page uncompressed streams:
>    - Keep a robust custom tokenizer that extracts strings/arrays preserving brackets `[...]` and parens `(...)`.
>    - Look for line drawing operators (`m` and `l`) that are horizontal and longer than 30 points.
>    - Look for rectangle drawing operators (`re`) that are typical text box sizes, supporting absolute mapping for negative heights.
>    - Track text position matrix adjustments (`Tm`, `Td`, `TD`) and font sizing (`Tf`).
>    - Detect consecutive underscore strings (`__________`) inside `Tj` or `TJ` text rendering arrays. Estimate text prefix widths dynamically to offset and position the interactive form overlays exactly over the underscores.
> 4. Implement a coordinate-based deduplication logic to avoid overlapping form fields.
> 5. For every detected blank geometry, map these to the `AcroForm` fields array by constructing a new `Annot` type `Widget` dictionary, and inject it as a text field.
> 6. Make sure the output is written as a new functional PDF (e.g. `/out/[input]_editable.pdf`).
> 7. Write a `main.go` inside `cmd/make-pdfs-editable-again/main.go` using standard `flag` argument parsing. Organize build binaries inside a `/bin/` directory.
> 8. Place the core processor code inside `pkg/pdfprocessor/processor.go`. Organize all test inputs and generated test forms under a dedicated `pkg/pdfprocessor/testdata/` directory.
> 9. Include a script in `samples/` that constructs a basic test PDF containing drawn lines, boxes, and text underscores using `github.com/jung-kurt/gofpdf` inside `pkg/pdfprocessor/testdata/test_form.pdf` so the tool logic can be tested locally without any API key restrictions.
> 10. Finally, write tests inside `pkg/pdfprocessor/processor_test.go` to automate this check and produce a comprehensive `README.md`.
