/**
 * Integration tests for the Go WASM build of lychee-layouts.
 *
 * Prerequisites:
 *   make build          (produces dist/lychee-layouts.wasm + dist/wasm_exec.js)
 *
 * Run:
 *   node --test tests/layouts.test.mjs
 *
 * Requires Node.js 18+.
 */

import { describe, test, before } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

// ── WASM bootstrap ────────────────────────────────────────────────────────────

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const distDir   = path.join(__dirname, '..', 'dist');

const wasmExecPath = path.join(distDir, 'wasm_exec.js');
const wasmBinPath  = path.join(distDir, 'lychee-layouts.wasm');

if (!existsSync(wasmExecPath) || !existsSync(wasmBinPath)) {
    console.error(
        '\nERROR: WASM artefacts not found.\n' +
        '  Run `make build` first, then re-run this test file.\n'
    );
    process.exit(1);
}

// Load the Go WASM runtime — populates globalThis.Go.
const _require = createRequire(import.meta.url);
_require(wasmExecPath);

const go = new globalThis.Go();
const wasmBuf = readFileSync(wasmBinPath);
const { instance } = await WebAssembly.instantiate(wasmBuf, go.importObject);
go.run(instance);

// Wait until Go main() registers lycheelayouts on globalThis.
await new Promise(resolve => {
    (function poll() {
        if (globalThis.lycheelayouts) return resolve();
        setTimeout(poll, 10);
    })();
});

const { justified, square, masonry, grid } = globalThis.lycheelayouts;

// ── Helpers ───────────────────────────────────────────────────────────────────

const EPS = 1e-9;
const near = (a, b) => Math.abs(a - b) < EPS;

/**
 * Assert all four fields of a box with human-readable error messages.
 * The JS WASM result uses lowercase keys: { top, left, width, height }.
 */
function assertBox(box, top, left, width, height) {
    assert.ok(near(box.top,    top),    `top:    ${box.top}    ≠ ${top}`);
    assert.ok(near(box.left,   left),   `left:   ${box.left}   ≠ ${left}`);
    assert.ok(near(box.width,  width),  `width:  ${box.width}  ≠ ${width}`);
    assert.ok(near(box.height, height), `height: ${box.height} ≠ ${height}`);
}

/** Returns an array of n copies of v. */
const repeat = (n, v) => Array.from({ length: n }, () => v);

// ════════════════════════════════════════════════════════════════════════════
// Square
// ════════════════════════════════════════════════════════════════════════════

describe('square', () => {
    test('empty_input_returns_default', () => {
        const r = square([], 300, 100, 0);
        assert.strictEqual(r.boxes.length, 0);
        assert.strictEqual(r.containerHeight, 0);
    });

    test('zero_container_width_returns_default', () => {
        const r = square([1, 1], 0, 100, 0);
        assert.strictEqual(r.boxes.length, 0);
    });

    test('zero_target_size_returns_default', () => {
        const r = square([1, 1], 300, 0, 0);
        assert.strictEqual(r.boxes.length, 0);
    });

    test('all_cells_are_square_regardless_of_ratio', () => {
        // cW=300, tS=100, gap=0 → cellSize=100; ratios must not influence cell size
        const r = square([2, 0.5, 3, 0.1, 1], 300, 100, 0);
        for (const [i, b] of r.boxes.entries()) {
            assert.ok(near(b.width, b.height), `box ${i}: not square (${b.width}×${b.height})`);
        }
    });

    test('single_photo_no_gap', () => {
        // cW=300, tS=100, gap=0 → perChunk=3, spread=0, cellSize=100
        const r = square([1], 300, 100, 0);
        assert.strictEqual(r.boxes.length, 1);
        assertBox(r.boxes[0], 0, 0, 100, 100);
        assert.ok(near(r.containerHeight, 100));
    });

    test('three_photos_fill_one_row_no_gap', () => {
        // cW=300, tS=100, gap=0 → cellSize=100, cols at 0, 100, 200
        const r = square([1, 1, 1], 300, 100, 0);
        assert.strictEqual(r.boxes.length, 3);
        assertBox(r.boxes[0], 0,   0, 100, 100);
        assertBox(r.boxes[1], 0, 100, 100, 100);
        assertBox(r.boxes[2], 0, 200, 100, 100);
        assert.ok(near(r.containerHeight, 100));
    });

    test('six_photos_two_rows_no_gap', () => {
        // cW=300, tS=100, gap=0 → 3 columns; row 0 at top=0, row 1 at top=100
        const r = square(repeat(6, 1), 300, 100, 0);
        for (let i = 0; i < 3; i++) assert.ok(near(r.boxes[i].top,   0), `photo ${i} should be in row 0`);
        for (let i = 3; i < 6; i++) assert.ok(near(r.boxes[i].top, 100), `photo ${i} should be in row 1`);
        assert.ok(near(r.containerHeight, 200));
    });

    test('spread_increases_cell_size_to_fill_container', () => {
        // cW=302, tS=100, gap=0 → perChunk=3, remaining=2, spread=ceil(2/3)=1, cellSize=101
        const r = square([1, 1, 1], 302, 100, 0);
        assertBox(r.boxes[0], 0,   0, 101, 101);
        assertBox(r.boxes[1], 0, 101, 101, 101);
        assertBox(r.boxes[2], 0, 202, 101, 101);
        assert.ok(near(r.containerHeight, 101));
    });

    test('gap_offsets_column_left_positions', () => {
        // cW=306, tS=100, gap=3 → perChunk=3, remaining=0, cellSize=100
        // col lefts: 0*(100+3)=0, 1*103=103, 2*103=206
        const r = square([1, 1, 1], 306, 100, 3);
        assertBox(r.boxes[0], 0,   0, 100, 100);
        assertBox(r.boxes[1], 0, 103, 100, 100);
        assertBox(r.boxes[2], 0, 206, 100, 100);
    });

    test('gap_offsets_row_top_positions', () => {
        // cW=306, tS=100, gap=3 → after row 0 col heights=103, row 1 at top=103
        const r = square(repeat(6, 1), 306, 100, 3);
        for (let i = 0; i < 3; i++) assert.ok(near(r.boxes[i].top,   0), `photo ${i} should be in row 0`);
        for (let i = 3; i < 6; i++) assert.ok(near(r.boxes[i].top, 103), `photo ${i} should be in row 1`);
        assert.ok(near(r.containerHeight, 206));
    });

    test('single_column_when_container_narrower_than_target', () => {
        // cW=50, tS=100, gap=0 → perChunk=max(1,0)=1, spread=-50, cellSize=50
        const r = square([1, 1], 50, 100, 0);
        assert.strictEqual(r.boxes.length, 2);
        assertBox(r.boxes[0],  0, 0, 50, 50);
        assertBox(r.boxes[1], 50, 0, 50, 50);
        assert.ok(near(r.containerHeight, 100));
    });
});

// ════════════════════════════════════════════════════════════════════════════
// Grid
// ════════════════════════════════════════════════════════════════════════════

describe('grid', () => {
    test('empty_input_returns_default', () => {
        const r = grid([], 300, 100, 0);
        assert.strictEqual(r.boxes.length, 0);
        assert.strictEqual(r.containerHeight, 0);
    });

    test('zero_container_width_returns_default', () => {
        const r = grid([1], 0, 100, 0);
        assert.strictEqual(r.boxes.length, 0);
    });

    test('photo_height_is_col_width_divided_by_ratio_floored', () => {
        // colWidth=100; h = floor(100 / ratio)
        const r = grid([2, 1, 0.5], 300, 100, 0);
        assert.strictEqual(r.boxes.length, 3);
        assertBox(r.boxes[0], 0,   0, 100,  50);  // floor(100/2.0) = 50
        assertBox(r.boxes[1], 0, 100, 100, 100);  // floor(100/1.0) = 100
        assertBox(r.boxes[2], 0, 200, 100, 200);  // floor(100/0.5) = 200
    });

    test('invalid_ratio_treated_as_one', () => {
        // ratio=0 or negative → falls back to 1 → h=floor(100/1)=100
        const r = grid([0, -2], 200, 100, 0);
        assert.ok(near(r.boxes[0].height, 100), 'box 0 height should be 100');
        assert.ok(near(r.boxes[1].height, 100), 'box 1 height should be 100');
    });

    test('all_photos_in_same_row_share_top', () => {
        // 3 photos into 3 columns → all in row 0, all at top=0
        const r = grid([2, 1, 0.5], 300, 100, 0);
        for (const [i, b] of r.boxes.entries()) {
            assert.ok(near(b.top, 0), `photo ${i}: top=${b.top}, want 0`);
        }
    });

    test('second_row_starts_at_tallest_column_of_first_row', () => {
        // Row 0: heights 50, 100, 200 → max=200 → row 1 syncs to top=200
        const r = grid([2, 1, 0.5, 1, 1, 1], 300, 100, 0);
        for (let i = 3; i < 6; i++) {
            assert.ok(near(r.boxes[i].top, 200), `photo ${i}: top=${r.boxes[i].top}, want 200`);
        }
    });

    test('second_row_correct_heights', () => {
        // Row 1: all ratio=1 → h=100 each; containerHeight = 200(sync) + 100 = 300
        const r = grid([2, 1, 0.5, 1, 1, 1], 300, 100, 0);
        assertBox(r.boxes[3], 200,   0, 100, 100);
        assertBox(r.boxes[4], 200, 100, 100, 100);
        assertBox(r.boxes[5], 200, 200, 100, 100);
        assert.ok(near(r.containerHeight, 300));
    });

    test('gap_offsets_row_and_column_positions', () => {
        // cW=306, tW=100, gap=3 → colWidth=100, col lefts: 0, 103, 206
        // After row 0: col heights=103, row 1 starts at top=103
        const r = grid(repeat(6, 1), 306, 100, 3);
        assertBox(r.boxes[0], 0,   0, 100, 100);
        assertBox(r.boxes[1], 0, 103, 100, 100);
        assertBox(r.boxes[2], 0, 206, 100, 100);
        for (let i = 3; i < 6; i++) {
            assert.ok(near(r.boxes[i].top, 103), `photo ${i}: top=${r.boxes[i].top}, want 103`);
        }
        assert.ok(near(r.containerHeight, 206));
    });
});

// ════════════════════════════════════════════════════════════════════════════
// Masonry
// ════════════════════════════════════════════════════════════════════════════

describe('masonry', () => {
    test('empty_input_returns_default', () => {
        const r = masonry([], 300, 100, 0);
        assert.strictEqual(r.boxes.length, 0);
        assert.strictEqual(r.containerHeight, 0);
    });

    test('zero_container_width_returns_default', () => {
        const r = masonry([1], 0, 100, 0);
        assert.strictEqual(r.boxes.length, 0);
    });

    test('equal_ratios_fill_columns_left_to_right', () => {
        // All columns start at 0; ties go to the lowest-index column (strict <).
        const r = masonry([1, 1, 1], 300, 100, 0);
        assert.strictEqual(r.boxes.length, 3);
        assertBox(r.boxes[0], 0,   0, 100, 100);
        assertBox(r.boxes[1], 0, 100, 100, 100);
        assertBox(r.boxes[2], 0, 200, 100, 100);
        assert.ok(near(r.containerHeight, 100));
    });

    test('portrait_photo_delays_its_column', () => {
        // Photo 0 (ratio=0.5) → col0 h=200; photos 1,2 → col1,col2; photo 3 → col1 (tie)
        const r = masonry([0.5, 1, 1, 1], 300, 100, 0);
        assertBox(r.boxes[0],   0,   0, 100, 200);
        assertBox(r.boxes[1],   0, 100, 100, 100);
        assertBox(r.boxes[2],   0, 200, 100, 100);
        assertBox(r.boxes[3], 100, 100, 100, 100); // stacked on top of photo 1
        assert.ok(near(r.containerHeight, 200));
    });

    test('landscape_photo_produces_short_cell', () => {
        // h = colWidth / ratio = 100 / 2 = 50
        const r = masonry([2, 1, 1], 300, 100, 0);
        assertBox(r.boxes[0], 0,   0, 100,  50); // wide → short cell
        assertBox(r.boxes[1], 0, 100, 100, 100);
        assertBox(r.boxes[2], 0, 200, 100, 100);
        assert.ok(near(r.containerHeight, 100));
    });

    test('next_photo_goes_to_shortest_column', () => {
        // Portrait in col0 (h=200) keeps it out of rotation until others grow.
        // Photo 3 (ratio=0.5) goes to col1 (tied with col2 at 100, first wins).
        const r = masonry([0.5, 1, 1, 0.5], 300, 100, 0);
        assert.ok(near(r.boxes[3].left,   100), `photo 3 left=${r.boxes[3].left}, want 100 (col1)`);
        assert.ok(near(r.boxes[3].top,    100), `photo 3 top=${r.boxes[3].top}, want 100`);
        assert.ok(near(r.boxes[3].height, 200), `photo 3 height=${r.boxes[3].height}, want 200`);
        assert.ok(near(r.containerHeight, 300));
    });

    test('invalid_ratio_treated_as_one', () => {
        const r = masonry([0, -1], 200, 100, 0);
        assert.ok(near(r.boxes[0].height, 100), 'box 0 height should be 100');
        assert.ok(near(r.boxes[1].height, 100), 'box 1 height should be 100');
    });

    test('gap_increases_column_heights', () => {
        // cW=306, tW=100, gap=3 → colWidth=100; each photo adds h+gap=103 to its column
        const r = masonry([1, 1, 1], 306, 100, 3);
        assert.ok(near(r.containerHeight, 103));
    });

    test('container_height_is_tallest_column', () => {
        // ratios=[1, 0.5, 1]: col0=100, col1=200, col2=100 → max=200
        const r = masonry([1, 0.5, 1], 300, 100, 0);
        assert.ok(near(r.containerHeight, 200));
    });
});

// ════════════════════════════════════════════════════════════════════════════
// Justified
// ════════════════════════════════════════════════════════════════════════════

describe('justified', () => {
    test('empty_input_returns_default', () => {
        const r = justified([], 300, 100, 0);
        assert.strictEqual(r.boxes.length, 0);
        assert.strictEqual(r.containerHeight, 0);
    });

    test('zero_container_width_returns_default', () => {
        const r = justified([1], 0, 100, 0);
        assert.strictEqual(r.boxes.length, 0);
    });

    test('non_positive_row_height_defaults_to_cw_over_four', () => {
        // tRH=0 → effective=400/4=100; single photo (ratio=1): ideal=400>100, last-row default h=100
        const r = justified([1], 400, 0, 0);
        assert.strictEqual(r.boxes.length, 1);
        assert.ok(near(r.boxes[0].height, 100));
    });

    test('single_landscape_photo_fills_full_row_width', () => {
        // ratio=2, cW=200, tRH=100: ideal=100=target → h=100, w=2×100=200
        const r = justified([2], 200, 100, 0);
        assertBox(r.boxes[0], 0, 0, 200, 100);
        assert.ok(near(r.containerHeight, 100));
    });

    test('single_portrait_photo_placed_alone', () => {
        // ratio=0.5, cW=200, tRH=100: ideal=400>target, last-row default h=100, w=0.5×100=50
        const r = justified([0.5], 200, 100, 0);
        assert.ok(near(r.boxes[0].height, 100));
        assert.ok(near(r.boxes[0].width,   50));
    });

    test('two_equal_photos_fill_row_exactly', () => {
        // ratios=[1,1], cW=200, tRH=100: j=1 ideal=100 → h=100, each w=100
        const r = justified([1, 1], 200, 100, 0);
        assertBox(r.boxes[0], 0,   0, 100, 100);
        assertBox(r.boxes[1], 0, 100, 100, 100);
        assert.ok(near(r.containerHeight, 100));
    });

    test('last_row_uses_target_height_without_stretching', () => {
        // 4 photos, cW=300, tRH=100: row0=3 photos h=100; row1=1 photo at h=100 (target, not stretched)
        const r = justified(repeat(4, 1), 300, 100, 0);
        for (let i = 0; i < 3; i++) assert.ok(near(r.boxes[i].top, 0), `photo ${i} should be in row 0`);
        assertBox(r.boxes[3], 100, 0, 100, 100);
        assert.ok(near(r.containerHeight, 200));
    });

    test('prefers_n_minus_one_photos_when_closer_to_target', () => {
        // ratios=[35,35,50], cW=8400, tRH=100:
        //   after j=1: ideal=120 (|120-100|=20)
        //   after j=2: ideal=70  (|70-100|=30)  → prefer j=1, row_height=120
        const r = justified([35, 35, 50], 8400, 100, 0);
        assert.strictEqual(r.boxes.length, 3);
        assertBox(r.boxes[0],   0,    0, 4200, 120);
        assertBox(r.boxes[1],   0, 4200, 4200, 120);
        assertBox(r.boxes[2], 120,    0, 5000, 100); // last row at target height
        assert.ok(near(r.containerHeight, 220));
    });

    test('prefers_n_photos_when_closer_to_target', () => {
        // ratios=[1,1,1], cW=300, tRH=100: j=2 ideal=100=target → use 3 photos
        const r = justified([1, 1, 1], 300, 100, 0);
        for (const b of r.boxes) assert.ok(near(b.height, 100), `height=${b.height}, want 100`);
        assert.ok(near(r.containerHeight, 100));
    });

    test('spacing_is_subtracted_from_available_width', () => {
        // ratios=[1,1], cW=202, tRH=100, spacing=2:
        //   j=1: avail=202-2=200, ideal=100 → h=100; photo 1 left=100+2=102
        const r = justified([1, 1], 202, 100, 2);
        assertBox(r.boxes[0], 0,   0, 100, 100);
        assertBox(r.boxes[1], 0, 102, 100, 100);
        assert.ok(near(r.containerHeight, 100));
    });

    test('spacing_advances_current_top_between_rows', () => {
        // 4 photos, cW=200, tRH=100, spacing=10:
        //   j=1: avail=190, ideal=95 → row_height=95; current_top = 95+10 = 105
        //   row 1 photos must start at top = row0.top + row0.height + spacing
        const r = justified(repeat(4, 1), 200, 100, 10);
        const row0H   = r.boxes[0].height;
        const row1Top = r.boxes[0].top + row0H + 10;
        assert.ok(near(r.boxes[2].top, row1Top), `photo 2 top=${r.boxes[2].top}, want ${row1Top}`);
        assert.ok(near(r.boxes[3].top, row1Top), `photo 3 top=${r.boxes[3].top}, want ${row1Top}`);
    });

    test('invalid_ratios_fall_back_to_one', () => {
        // ratios=[0,-1] treated as [1,1] → same as two_equal_photos_fill_row_exactly
        const r = justified([0, -1], 200, 100, 0);
        assert.ok(near(r.boxes[0].height, 100));
        assert.ok(near(r.boxes[1].height, 100));
        assert.ok(near(r.containerHeight, 100));
    });
});
