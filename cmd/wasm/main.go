//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/aawc/make-pdfs-editable-again/pkg/pdfprocessor"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func makePdfEditableWrapper(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{"error": "missing PDF bytes argument"}
	}

	// 1. Retrieve Uint8Array from JS
	jsBytes := args[0]
	length := jsBytes.Get("byteLength").Int()
	inBytes := make([]byte, length)
	js.CopyBytesToGo(inBytes, jsBytes)

	// 2. Run our FOSS core processing engine in-memory
	outBytes, err := pdfprocessor.DetectAndInjectBytes(inBytes)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	// 3. Copy the processed output bytes back to JS Uint8Array
	uint8ArrayClass := js.Global().Get("Uint8Array")
	jsOutBytes := uint8ArrayClass.New(len(outBytes))
	js.CopyBytesToJS(jsOutBytes, outBytes)

	return map[string]interface{}{
		"pdfBytes": jsOutBytes,
	}
}

func main() {
	api.DisableConfigDir()
	c := make(chan struct{}, 0)
	fmt.Println("PDF Form Filler WebAssembly initialized successfully.")
	// Expose the Go function globally to JS
	js.Global().Set("makePdfEditable", js.FuncOf(makePdfEditableWrapper))
	<-c
}
