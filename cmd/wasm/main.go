//go:build js && wasm

// Command wasm exposes srt2fcpxml to the web page as the global object
// srt2fcpxml with two functions; files never leave the browser:
//
//	srt2fcpxml.convert(requestJSON, data, secondaryData) → responseJSON
//	srt2fcpxml.extract(requestJSON, data) → responseJSON
//
// data and secondaryData are Uint8Arrays (secondaryData may be null). See
// internal/webapi for the request and response fields.
package main

import (
	"syscall/js"

	"github.com/GanymedeNil/srt2fcpxml/v3/internal/webapi"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func bytesArg(args []js.Value, i int) []byte {
	if i >= len(args) || args[i].IsNull() || args[i].IsUndefined() {
		return nil
	}
	b := make([]byte, args[i].Get("length").Int())
	js.CopyBytesToGo(b, args[i])
	return b
}

func stringArg(args []js.Value, i int) string {
	if i >= len(args) || args[i].Type() != js.TypeString {
		return ""
	}
	return args[i].String()
}

func main() {
	api := js.Global().Get("Object").New()
	api.Set("version", version)
	api.Set("convert", js.FuncOf(func(_ js.Value, args []js.Value) any {
		return webapi.Convert(stringArg(args, 0), bytesArg(args, 1), bytesArg(args, 2))
	}))
	api.Set("extract", js.FuncOf(func(_ js.Value, args []js.Value) any {
		return webapi.Extract(stringArg(args, 0), bytesArg(args, 1))
	}))
	js.Global().Set("srt2fcpxml", api)
	js.Global().Call("dispatchEvent", js.Global().Get("Event").New("srt2fcpxml-ready"))
	select {} // keep the functions alive
}
