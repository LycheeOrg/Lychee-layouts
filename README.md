# Lychee-layouts

[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/LycheeOrg/Lychee-layouts/badge)](https://securityscorecards.dev/viewer/?uri=github.com/LycheeOrg/Lychee-layouts)

Photo layout engine for [Lychee](https://lycheeorg.github.io), compiled to WebAssembly.

Provides four layout algorithms — **justified**, **masonry**, **grid**, and **square** — as a pure-logic library with no DOM dependency. It ships in two implementations that produce identical results:

| Implementation | Toolchain | Typical output size |
|---|---|---|
| **Go** (`/`) | `go build` / TinyGo | ~2 MB / ~400 KB |
| **Rust** (`rust/`) | `wasm-pack` | ~80 KB |

---

## Layout algorithms

All four algorithms take a list of **aspect ratios** (width ÷ height) and container parameters, and return a flat list of **boxes** — one per photo — with `top`, `left`, `width`, and `height` in pixels, plus the required `containerHeight`.

### Justified

Every photo in a row shares the same height and the row fills the container width exactly. The last (partial) row is placed at `targetRowHeight` without stretching.

```
┌──────┐ ┌────────────┐ ┌─────┐
│      │ │            │ │     │  ← same height, fills width
└──────┘ └────────────┘ └─────┘
┌───────────┐ ┌──────┐ ┌─────┐
│           │ │      │ │     │  ← same height, fills width
└───────────┘ └──────┘ └─────┘
┌──────┐ ┌──────┐                ← last row: targetRowHeight, not stretched
│      │ │      │
└──────┘ └──────┘
```

> Replaces the [`justified-layout`](https://github.com/flickr/justified-layout) npm package with a zero-dependency Go/Rust implementation of the same greedy algorithm.

### Masonry

Each photo flows into the shortest column (Pinterest-style). Columns are not row-aligned.

```
┌─────┐  ┌─────┐  ┌─────┐
│     │  │     │  │     │
│     │  └─────┘  │     │
└─────┘  ┌─────┐  └─────┘
┌─────┐  │     │  ┌─────┐
│     │  │     │  │     │
└─────┘  └─────┘  └─────┘
```

### Grid

Fixed columns, variable photo height per cell. All columns are synced to the same `top` at the start of each row, so rows stay horizontally aligned even though individual photo heights differ.

```
┌─────┐  ┌─────┐  ┌─────┐  ← row start aligned
│     │  │     │  │     │
│     │  └─────┘  └─────┘
└─────┘
┌─────┐  ┌─────┐  ┌─────┐  ← next row starts at max bottom of previous
│     │  │     │  │     │
└─────┘  └─────┘  │     │
                  └─────┘
```

### Square

Every photo is rendered as a square cell (`targetSize × targetSize`). Aspect ratios are ignored. Cells are distributed evenly to fill the container width.

```
┌─────┐  ┌─────┐  ┌─────┐
│     │  │     │  │     │
└─────┘  └─────┘  └─────┘
┌─────┐  ┌─────┐  ┌─────┐
│     │  │     │  │     │
└─────┘  └─────┘  └─────┘
```

---

## API

All four functions share the same signature pattern and return the same shape.

### Input

| Parameter | Type | Description |
|---|---|---|
| `ratios` | `Float64Array` | Aspect ratio (`width / height`) for each photo. Falls back to `1.0` for invalid values. |
| `containerWidth` | `f64` / `number` | Usable container width in pixels (exclude padding and scrollbar). |
| `targetRowHeight` *(justified)* | `f64` / `number` | Desired row height in pixels. |
| `targetSize` *(square)* | `f64` / `number` | Desired cell size in pixels. |
| `targetWidth` *(masonry, grid)* | `f64` / `number` | Desired column width in pixels. |
| `gap` / `spacing` | `f64` / `number` | Spacing between photos in pixels (horizontal and vertical). |

**Computing ratios** from photo metadata:
```js
const ratios = photos.map(p => p.width > 0 && p.height > 0 ? p.width / p.height : 1.0);
```

### Output

```ts
type Box = { top: number; left: number; width: number; height: number };
type LayoutResult = { containerHeight: number; boxes: Box[] };
```

`boxes[i]` corresponds to `ratios[i]`. Set `containerHeight` on your container element so that absolutely-positioned children are not clipped.

---

## Building

### Rust (recommended — smallest output)

**Prerequisites:**
```bash
curl https://sh.rustup.rs -sSf | sh        # install Rust
cargo install wasm-pack                    # install wasm-pack
```

**Build:**
```bash
make build-rs
# output: dist/rs/lychee_layouts_bg.wasm + JS/TS glue
```

Or manually from the `rust/` directory:
```bash
cd rust
wasm-pack build --target web --out-dir ../dist/rs --release
```

### Go

**Prerequisites:** Go 1.21+

```bash
make build
# output: dist/Lychee-layouts.wasm + dist/wasm_exec.js
```

For a smaller binary (~5× smaller) with [TinyGo](https://tinygo.org):
```bash
make build-tinygo
# output: dist/Lychee-layouts-tiny.wasm
```

---

## Usage

### Rust WASM (wasm-pack output)

`wasm-pack` generates a JavaScript/TypeScript module alongside the `.wasm` file. Import it as a standard ES module:

```html
<script type="module">
import init, { justified, square, masonry, grid } from './dist/rs/lychee_layouts.js';

await init(); // load and compile the WASM module once

const ratios = new Float64Array(photos.map(p => p.width / p.height));
const containerWidth = document.getElementById('gallery').clientWidth;

// Justified layout
const { containerHeight, boxes } = justified(ratios, containerWidth, 320, 3);

applyLayout(containerHeight, boxes);
</script>
```

TypeScript types are generated automatically by `wasm-pack` in `dist/rs/lychee_layouts.d.ts`.

### Go WASM

The Go WASM module exposes a `lycheelayouts` global object. It requires `wasm_exec.js` (shipped with Go) as a loader.

```html
<script src="dist/wasm_exec.js"></script>
<script>
const go = new Go();
WebAssembly.instantiateStreaming(fetch('dist/Lychee-layouts.wasm'), go.importObject)
  .then(({ instance }) => {
    go.run(instance);

    const ratios = photos.map(p => p.width / p.height);
    const containerWidth = document.getElementById('gallery').clientWidth;

    // Justified layout
    const { containerHeight, boxes } = lycheelayouts.justified(
      ratios, containerWidth, 320, 3
    );

    applyLayout(containerHeight, boxes);
  });
</script>
```

### Applying the result

```js
function applyLayout(containerHeight, boxes) {
  const el = document.getElementById('gallery');
  el.style.position = 'relative';
  el.style.height = containerHeight + 'px';

  const items = el.querySelectorAll('.photo-item');
  boxes.forEach((box, i) => {
    const item = items[i];
    item.style.position = 'absolute';
    item.style.top    = box.top    + 'px';
    item.style.left   = box.left   + 'px';
    item.style.width  = box.width  + 'px';
    item.style.height = box.height + 'px';
  });
}
```

Call `applyLayout` on page load and whenever the container is resized (debounce as needed):

```js
const resizeObserver = new ResizeObserver(() => {
  const { containerHeight, boxes } = justified(ratios, el.clientWidth, 320, 3);
  applyLayout(containerHeight, boxes);
});
resizeObserver.observe(document.getElementById('gallery'));
```

---

## Function reference

### `justified(ratios, containerWidth, targetRowHeight, spacing)`

```
ratios          Float64Array   aspect ratios of all photos
containerWidth  number         usable container width in px
targetRowHeight number         desired row height in px  (default: containerWidth / 4)
spacing         number         gap between photos in px
```

Implements the [Flickr justified-layout](https://github.com/flickr/justified-layout) greedy algorithm without the npm dependency. The last incomplete row is placed at `targetRowHeight` (not stretched).

---

### `square(ratios, containerWidth, targetSize, gap)`

```
ratios          Float64Array   aspect ratios — used only for the item count
containerWidth  number         usable container width in px
targetSize      number         desired square cell size in px
gap             number         gap between cells in px
```

Distributes cells evenly: `cellSize = targetSize + ceil(remainingSpace / columns)`.

---

### `masonry(ratios, containerWidth, targetWidth, gap)`

```
ratios          Float64Array   aspect ratios of all photos
containerWidth  number         usable container width in px
targetWidth     number         desired column width in px
gap             number         gap between photos in px
```

Each photo's height is `columnWidth / ratio`. New photos always go into the column with the smallest accumulated height.

---

### `grid(ratios, containerWidth, targetWidth, gap)`

```
ratios          Float64Array   aspect ratios of all photos
containerWidth  number         usable container width in px
targetWidth     number         desired column width in px
gap             number         gap between photos in px
```

Like masonry but row-aligned: at the start of every new row all column heights are synced to the tallest column, so rows begin at a consistent horizontal baseline.

---

## Project structure

```
.
├── Makefile
│
├── layouts/              Go library (pure, no WASM dependency)
│   ├── types.go          Box, LayoutResult
│   ├── justified.go
│   ├── masonry.go
│   ├── grid.go
│   └── square.go
│
├── cmd/wasm/
│   └── main.go           Go WASM entry point (syscall/js)
│
├── rust/                 Rust implementation
│   ├── Cargo.toml
│   └── src/
│       ├── lib.rs        wasm-bindgen exports
│       ├── types.rs      PhotoBox, LayoutResult
│       ├── justified.rs
│       ├── masonry.rs
│       ├── grid.rs
│       └── square.rs
│
└── ref/                  Original TypeScript reference (Lychee source)
    ├── PhotoLayout.ts
    ├── useJustify.ts
    ├── useMasonry.ts
    ├── useGrid.ts
    └── useSquare.ts
```

---

## Make targets

| Target | Description |
|---|---|
| `make build` | Go WASM (standard toolchain) → `dist/` |
| `make build-tinygo` | Go WASM (TinyGo, smaller) → `dist/` |
| `make build-rs` | Rust WASM (wasm-pack) → `dist/rs/` |
| `make test` | Go unit tests |
| `make test-rs` | Rust unit tests (native, no WASM runtime needed) |
| `make clean` | Remove `dist/` and Rust build artifacts |

---

## License

Part of the [Lychee](https://github.com/LycheeOrg/Lychee) project.
