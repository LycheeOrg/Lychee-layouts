//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/LycheeOrg/lychee-layouts/layouts"
)

func main() {
	js.Global().Set("lycheelayouts", js.ValueOf(map[string]any{
		"justified": js.FuncOf(justified),
		"square":    js.FuncOf(square),
		"masonry":   js.FuncOf(masonry),
		"grid":      js.FuncOf(grid),
	}))

	// Keep the WASM module alive indefinitely.
	select {}
}

// jsArrayToFloat64 converts a JS Array or TypedArray to a Go slice.
func jsArrayToFloat64(v js.Value) []float64 {
	n := v.Length()
	out := make([]float64, n)
	for i := range n {
		out[i] = v.Index(i).Float()
	}
	return out
}

// toJS serialises a LayoutResult to a JS object via JSON round-trip.
func toJS(r layouts.LayoutResult) js.Value {
	b, err := json.Marshal(r)
	if err != nil {
		return js.Null()
	}
	return js.Global().Get("JSON").Call("parse", string(b))
}

// justified(ratios, containerWidth, targetRowHeight, spacing) → {containerHeight, boxes}
func justified(this js.Value, args []js.Value) any {
	if len(args) < 4 {
		return js.Null()
	}
	result := layouts.Justified(
		jsArrayToFloat64(args[0]),
		args[1].Float(),
		args[2].Float(),
		args[3].Float(),
	)
	return toJS(result)
}

// square(ratios, containerWidth, targetSize, gap) → {containerHeight, boxes}
func square(this js.Value, args []js.Value) any {
	if len(args) < 4 {
		return js.Null()
	}
	result := layouts.Square(
		jsArrayToFloat64(args[0]),
		args[1].Float(),
		args[2].Float(),
		args[3].Float(),
	)
	return toJS(result)
}

// masonry(ratios, containerWidth, targetWidth, gap) → {containerHeight, boxes}
func masonry(this js.Value, args []js.Value) any {
	if len(args) < 4 {
		return js.Null()
	}
	result := layouts.Masonry(
		jsArrayToFloat64(args[0]),
		args[1].Float(),
		args[2].Float(),
		args[3].Float(),
	)
	return toJS(result)
}

// grid(ratios, containerWidth, targetWidth, gap) → {containerHeight, boxes}
func grid(this js.Value, args []js.Value) any {
	if len(args) < 4 {
		return js.Null()
	}
	result := layouts.Grid(
		jsArrayToFloat64(args[0]),
		args[1].Float(),
		args[2].Float(),
		args[3].Float(),
	)
	return toJS(result)
}
