package pdfprocessor

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// Blank represents a visual line or rectangle where a user might type.
type Blank struct {
	PageNum     int
	X, Y        float64
	Width       float64
	Height      float64
	IsRectangle bool
}

// TextBox represents a visual text element's bounding area
type TextBox struct {
	MinX, MinY, MaxX, MaxY float64
}

func DetectAndInject(inputPath, outputPath string) error {
	inBytes, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	outBytes, err := DetectAndInjectBytes(inBytes)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, outBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}

func DetectAndInjectBytes(inBytes []byte) ([]byte, error) {
	rs := bytes.NewReader(inBytes)
	ctx, err := api.ReadAndValidate(rs, model.NewDefaultConfiguration())
	if err != nil {
		return nil, fmt.Errorf("could not read or validate pdf context: %v", err)
	}

	if err := api.OptimizeContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to optimize context: %v", err)
	}

	var allBlanks []Blank
	pageCount := ctx.PageCount

	for i := 1; i <= pageCount; i++ {
		blanks, err := extractBlanksFromPage(ctx, i)
		if err != nil {
			return nil, err
		}
		allBlanks = append(allBlanks, blanks...)
	}

	allBlanks = deduplicateBlanks(allBlanks)

	fmt.Printf("Detected %d blanks across %d pages.\n", len(allBlanks), pageCount)

	// Inject Text fields
	err = addFormFields(ctx, allBlanks)
	if err != nil {
		return nil, fmt.Errorf("failed to add form fields: %v", err)
	}

	var outBuf bytes.Buffer
	err = api.WriteContext(ctx, &outBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to write output pdf: %v", err)
	}

	return outBuf.Bytes(), nil
}


func extractBlanksFromPage(ctx *model.Context, pageNum int) ([]Blank, error) {
	var blanks []Blank
	var textBoxes []TextBox

	// Note: in pdfcpu v0.6.x+ to get page content:
	// Use ctx.PageContent(pageDict, pageObjNr)
	pageDict, pageDictIndRef, _, err := ctx.PageDict(pageNum, false)
	if err != nil {
		return nil, err
	}

	streamBytes, err := ctx.PageContent(pageDict, pageDictIndRef.ObjectNumber.Value())
	if err != nil || len(streamBytes) == 0 {
		return nil, nil // Some pages may have empty stream
	}

	tokens := tokenize(streamBytes)

	var lastX, lastY float64
	var textX, textY float64
	var lineX, lineY float64
	var lastFontSize float64 = 11.0

	underscoreRe := regexp.MustCompile(`_+`)

	for i := 0; i < len(tokens); i++ {
		op := string(tokens[i])

		switch op {
		case "m": // move to
			if i >= 2 {
				lastX = parseNum(string(tokens[i-2]))
				lastY = parseNum(string(tokens[i-1]))
			}
		case "l": // line to
			if i >= 2 {
				toX := parseNum(string(tokens[i-2]))
				toY := parseNum(string(tokens[i-1]))

				diffY := toY - lastY
				if diffY < 0 {
					diffY = -diffY
				}

				if diffY < 2.0 {
					length := toX - lastX
					if length < 0 {
						length = -length
						// swap
						tmp := toX
						toX = lastX
						lastX = tmp
					}

					if length > 30.0 {
						blanks = append(blanks, Blank{
							PageNum:     pageNum,
							X:           lastX,
							Y:           lastY,
							Width:       length,
							Height:      1,
							IsRectangle: false,
						})
					}
				}
				lastX = toX
				lastY = toY
			}
		case "re": // rectangle
			if i >= 4 {
				x := parseNum(string(tokens[i-4]))
				y := parseNum(string(tokens[i-3]))
				w := parseNum(string(tokens[i-2]))
				h := parseNum(string(tokens[i-1]))

				wAbs := w
				if wAbs < 0 {
					wAbs = -wAbs
				}
				hAbs := h
				if hAbs < 0 {
					hAbs = -hAbs
				}

				if wAbs > 30 && wAbs < 600 && hAbs > 12 && hAbs < 400 {
					rectX := x
					rectY := y
					if w < 0 {
						rectX = x + w
					}
					if h < 0 {
						rectY = y + h
					}

					blanks = append(blanks, Blank{
						PageNum:     pageNum,
						X:           rectX,
						Y:           rectY,
						Width:       wAbs,
						Height:      hAbs,
						IsRectangle: true,
					})
				}
			}
		case "BT":
			textX = 0
			textY = 0
			lineX = 0
			lineY = 0
		case "Td":
			if i >= 2 {
				tx := parseNum(string(tokens[i-2]))
				ty := parseNum(string(tokens[i-1]))
				lineX += tx
				lineY += ty
				textX = lineX
				textY = lineY
			}
		case "TD":
			if i >= 2 {
				tx := parseNum(string(tokens[i-2]))
				ty := parseNum(string(tokens[i-1]))
				lineX += tx
				lineY += ty
				textX = lineX
				textY = lineY
			}
		case "Tm":
			if i >= 6 {
				e := parseNum(string(tokens[i-2]))
				f := parseNum(string(tokens[i-1]))
				lineX = e
				lineY = f
				textX = lineX
				textY = lineY
			}
		case "Tf":
			if i >= 2 {
				lastFontSize = parseNum(string(tokens[i-1]))
			}
		case "TJ", "Tj":
			if i >= 1 {
				arg := string(tokens[i-1])
				cleanStr := cleanTJString(arg)

				// 1. If it contains alphanumeric characters (non-underscore labels),
				// save its bounding box to filter out overlapping decorative rectangles.
				stripped := strings.ReplaceAll(cleanStr, "_", "")
				stripped = strings.TrimSpace(stripped)
				if len(stripped) > 0 {
					w := estimateTextWidth(cleanStr, lastFontSize)
					textBoxes = append(textBoxes, TextBox{
						MinX: textX,
						MinY: textY,
						MaxX: textX + w,
						MaxY: textY + lastFontSize,
					})
				}

				matches := underscoreRe.FindAllString(arg, -1)
				totalUnderscores := 0
				for _, m := range matches {
					totalUnderscores += len(m)
				}
				if totalUnderscores > 5 {
					firstUnderscoreIdx := strings.Index(cleanStr, "_")
					if firstUnderscoreIdx == -1 {
						firstUnderscoreIdx = 0
					}

					// Estimate prefix width dynamically
					prefixChars := cleanStr[:firstUnderscoreIdx]
					prefixWidth := estimateTextWidth(prefixChars, lastFontSize)

					width := estimateTextWidth(strings.Repeat("_", totalUnderscores), lastFontSize)
					fieldX := textX + prefixWidth
					fieldY := textY

					height := lastFontSize
					if height < 10 {
						height = 10
					}

					blanks = append(blanks, Blank{
						PageNum:     pageNum,
						X:           fieldX,
						Y:           fieldY,
						Width:       width,
						Height:      height,
						IsRectangle: false,
					})
				}
			}
		}
	}

	// 2. Filter out any detected rectangles that contain pre-existing text elements
	var filteredBlanks []Blank
	for _, b := range blanks {
		if b.IsRectangle {
			intersects := false
			for _, t := range textBoxes {
				// AABB intersection checks: skip if not overlapping
				if !(t.MaxX < b.X || t.MinX > b.X+b.Width || t.MaxY < b.Y || t.MinY > b.Y+b.Height) {
					intersects = true
					break
				}
			}
			if intersects {
				// Discard false-positive table boundaries or legal declaration boxes
				continue
			}
		}
		filteredBlanks = append(filteredBlanks, b)
	}

	return filteredBlanks, nil
}

func tokenize(data []byte) []string {
	var tokens []string
	var current bytes.Buffer
	inParens := 0
	inBrackets := 0
	escaped := false

	for i := 0; i < len(data); i++ {
		c := data[i]

		if inParens > 0 {
			current.WriteByte(c)
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == '(' {
				inParens++
			} else if c == ')' {
				inParens--
			}
			continue
		}

		if c == '(' {
			inParens = 1
			current.WriteByte(c)
			continue
		}

		if inBrackets > 0 {
			current.WriteByte(c)
			if c == '[' {
				inBrackets++
			} else if c == ']' {
				inBrackets--
				if inBrackets == 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
			}
			continue
		}

		if c == '[' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			inBrackets = 1
			current.WriteByte(c)
			continue
		}

		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(c)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func cleanTJString(s string) string {
	var buf bytes.Buffer
	inParens := false
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			buf.WriteByte(c)
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '(' {
			inParens = true
			continue
		}
		if c == ')' {
			inParens = false
			continue
		}
		if inParens {
			buf.WriteByte(c)
		}
	}
	return buf.String()
}

func parseNum(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func estimateTextWidth(s string, fontSize float64) float64 {
	var width float64
	for _, r := range s {
		wFactor := 0.53 // default
		switch {
		case r == '_':
			wFactor = 0.50 // Precise Helvetica underscore factor
		case r >= 'A' && r <= 'Z' || r == 'm' || r == 'w' || r == '@' || r == '%':
			wFactor = 0.7
		case strings.ContainsRune("iltfjrI ():.,-", r):
			wFactor = 0.3
		}
		width += wFactor * fontSize
	}
	return width
}

func addFormFields(ctx *model.Context, blanks []Blank) error {
	rootDict := ctx.RootDict
	if rootDict == nil {
		return fmt.Errorf("no root dict found")
	}

	// 1. Register Helvetica Font for AcroForm Default Resources
	fontDict := types.NewDict()
	fontDict.Insert("Type", types.Name("Font"))
	fontDict.Insert("Subtype", types.Name("Type1"))
	fontDict.Insert("BaseFont", types.Name("Helvetica"))
	fontDict.Insert("Encoding", types.Name("WinAnsiEncoding"))

	fontIndRef, err := ctx.IndRefForNewObject(fontDict)
	if err != nil {
		return fmt.Errorf("failed to register Helv font: %v", err)
	}

	fontResources := types.NewDict()
	fontResources.Insert("Helv", *fontIndRef)

	drDict := types.NewDict()
	drDict.Insert("Font", fontResources)

	// 2. Retrieve or Create AcroForm
	var acroForm types.Dict
	acroFormObj, found := rootDict.Find("AcroForm")
	if found {
		acroFormDict, err := ctx.DereferenceDict(acroFormObj)
		if err == nil && acroFormDict != nil {
			acroForm = acroFormDict
		}
	} else {
		acroForm = types.NewDict()
		indRef, err := ctx.IndRefForNewObject(acroForm)
		if err != nil {
			return err
		}
		rootDict.Insert("AcroForm", *indRef)
	}

	// 3. Set required AcroForm properties
	acroForm.Insert("NeedAppearances", types.Boolean(true))
	acroForm.Insert("DR", drDict)
	acroForm.Insert("DA", types.StringLiteral("/Helv 0 Tf 0 g"))

	var fields types.Array
	fieldsObj, found := acroForm.Find("Fields")
	if found {
		fArray, err := ctx.DereferenceArray(fieldsObj)
		if err == nil && fArray != nil {
			fields = fArray
		}
	} else {
		fields = types.Array{}
	}

	fieldCount := len(fields) + 1

	pageDicts := make(map[int]types.Dict)

	for _, blank := range blanks {
		h := blank.Height
		if h < 5 {
			h = 15
		}
		w := blank.Width

		llx := blank.X
		lly := blank.Y
		urx := blank.X + w
		ury := lly + h

		rectArray := types.Array{
			types.Float(llx),
			types.Float(lly),
			types.Float(urx),
			types.Float(ury),
		}

		fieldDict := types.NewDict()
		fieldDict.Insert("Type", types.Name("Annot"))
		fieldDict.Insert("Subtype", types.Name("Widget"))
		fieldDict.Insert("FT", types.Name("Tx"))
		// Print flag
		fieldDict.Insert("F", types.Integer(4))
		fieldDict.Insert("T", types.StringLiteral(fmt.Sprintf("TextField_%d", fieldCount)))
		fieldDict.Insert("Rect", rectArray)
		fieldDict.Insert("DA", types.StringLiteral("/Helv 0 Tf 0 g")) // Set local DA for Chrome compatibility

		fieldIndRef, err := ctx.IndRefForNewObject(fieldDict)
		if err != nil {
			return err
		}
		fields = append(fields, *fieldIndRef)

		// Retrieve or use cached Page Dict
		pageDict, ok := pageDicts[blank.PageNum]
		if !ok {
			var err error
			pageDict, _, _, err = ctx.PageDict(blank.PageNum, false)
			if err != nil {
				return fmt.Errorf("failed to get page dict for page %d: %v", blank.PageNum, err)
			}
			pageDicts[blank.PageNum] = pageDict
		}

		var annots types.Array
		annotsObj, found := pageDict.Find("Annots")
		if found {
			if a, ok := annotsObj.(types.Array); ok {
				annots = a
			} else if ir, ok := annotsObj.(types.IndirectRef); ok {
				aArr, err := ctx.DereferenceArray(ir)
				if err == nil && aArr != nil {
					annots = aArr
				}
			} else {
				aArr, err := ctx.DereferenceArray(annotsObj)
				if err == nil && aArr != nil {
					annots = aArr
				}
			}
		}
		annots = append(annots, *fieldIndRef)
		pageDict.Update("Annots", annots)

		fieldCount++
	}

	acroForm.Update("Fields", fields)
	return nil
}

func deduplicateBlanks(blanks []Blank) []Blank {
	var unique []Blank
	for _, b := range blanks {
		isDuplicate := false
		for _, u := range unique {
			if u.PageNum == b.PageNum &&
				math.Abs(u.X-b.X) < 2.0 &&
				math.Abs(u.Y-b.Y) < 2.0 &&
				math.Abs(u.Width-b.Width) < 5.0 &&
				u.IsRectangle == b.IsRectangle {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			unique = append(unique, b)
		}
	}
	return unique
}
