WASM_OUT    := dist/lychee-layouts.wasm
WASM_EXEC   := dist/wasm_exec.js
RS_OUT_DIR  := dist/rs

.PHONY: build build-tinygo build-rs test test-js test-rs clean

# ── Go ────────────────────────────────────────────────────────────────────────

## Build with standard Go toolchain
build:
	mkdir -p dist
	GOOS=js GOARCH=wasm go build -o $(WASM_OUT) ./cmd/wasm
	cp "$(shell go env GOROOT)/lib/wasm/wasm_exec.js" $(WASM_EXEC) 2>/dev/null || \
	  cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" $(WASM_EXEC)
	@echo "Built $(WASM_OUT) ($$(du -sh $(WASM_OUT) | cut -f1))"

## Build with TinyGo for a smaller Go output (~5× smaller)
build-tinygo:
	mkdir -p dist
	tinygo build -o dist/lychee-layouts-tiny.wasm -target wasm ./cmd/wasm
	cp "$(shell tinygo env TINYGOROOT)/targets/wasm_exec.js" dist/wasm_exec_tiny.js
	@echo "Built dist/lychee-layouts-tiny.wasm ($$(du -sh dist/lychee-layouts-tiny.wasm | cut -f1))"

## Run unit tests for the Go layouts package
test:
	go test ./layouts/...

## Run JavaScript tests against the Go WASM build (requires: make build)
test-js:
	node --test tests/layouts.test.mjs

# ── Rust ──────────────────────────────────────────────────────────────────────

## Build with wasm-pack (installs deps on first run: cargo install wasm-pack)
## Output: dist/rs/lychee_layouts{_bg}.wasm + JS glue
build-rs:
	mkdir -p $(RS_OUT_DIR)
	cd rust && wasm-pack build --target web --out-dir ../$(RS_OUT_DIR) --release
	@echo "Built $(RS_OUT_DIR)/lychee_layouts_bg.wasm ($$(du -sh $(RS_OUT_DIR)/lychee_layouts_bg.wasm | cut -f1))"

## Run Rust unit tests (native target, no WASM runtime needed)
test-rs:
	cd rust && cargo test

# ── Shared ───────────────────────────────────────────────────────────────────

clean:
	rm -rf dist/
	cd rust && cargo clean
