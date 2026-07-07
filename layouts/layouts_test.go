package layouts_test

import (
	"math"
	"testing"

	"github.com/LycheeOrg/lychee-layouts/layouts"
)

// ── helpers ───────────────────────────────────────────────────────────────────

const eps = 1e-9

func near(a, b float64) bool { return math.Abs(a-b) < eps }

func checkBox(t *testing.T, got layouts.Box, top, left, width, height float64) {
	t.Helper()
	if !near(got.Top, top) {
		t.Errorf("  top:    got %.6g, want %.6g", got.Top, top)
	}
	if !near(got.Left, left) {
		t.Errorf("  left:   got %.6g, want %.6g", got.Left, left)
	}
	if !near(got.Width, width) {
		t.Errorf("  width:  got %.6g, want %.6g", got.Width, width)
	}
	if !near(got.Height, height) {
		t.Errorf("  height: got %.6g, want %.6g", got.Height, height)
	}
}

// repeat returns a slice of n copies of v.
func repeat(n int, v float64) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = v
	}
	return s
}

// ════════════════════════════════════════════════════════════════════════════
// Square
// ════════════════════════════════════════════════════════════════════════════

func TestSquare(t *testing.T) {
	t.Run("empty_input_returns_default", func(t *testing.T) {
		r := layouts.Square(nil, 300, 100, 0)
		if len(r.Boxes) != 0 {
			t.Fatalf("expected 0 boxes, got %d", len(r.Boxes))
		}
		if r.ContainerHeight != 0 {
			t.Errorf("containerHeight: got %v, want 0", r.ContainerHeight)
		}
	})

	t.Run("zero_container_width_returns_default", func(t *testing.T) {
		r := layouts.Square([]float64{1, 1}, 0, 100, 0)
		if len(r.Boxes) != 0 {
			t.Fatalf("expected 0 boxes, got %d", len(r.Boxes))
		}
	})

	t.Run("zero_target_size_returns_default", func(t *testing.T) {
		r := layouts.Square([]float64{1, 1}, 300, 0, 0)
		if len(r.Boxes) != 0 {
			t.Fatalf("expected 0 boxes, got %d", len(r.Boxes))
		}
	})

	t.Run("all_cells_are_square_regardless_of_ratio", func(t *testing.T) {
		r := layouts.Square([]float64{2, 0.5, 3, 0.1, 1}, 300, 100, 0)
		for i, b := range r.Boxes {
			if !near(b.Width, b.Height) {
				t.Errorf("box %d: not square (%.6g × %.6g)", i, b.Width, b.Height)
			}
		}
	})

	t.Run("single_photo_no_gap", func(t *testing.T) {
		// cW=300, tS=100, gap=0 → perChunk=3, spread=0, cellSize=100
		r := layouts.Square([]float64{1}, 300, 100, 0)
		if len(r.Boxes) != 1 {
			t.Fatalf("expected 1 box, got %d", len(r.Boxes))
		}
		checkBox(t, r.Boxes[0], 0, 0, 100, 100)
		if !near(r.ContainerHeight, 100) {
			t.Errorf("containerHeight: got %v, want 100", r.ContainerHeight)
		}
	})

	t.Run("three_photos_fill_one_row_no_gap", func(t *testing.T) {
		// cW=300, tS=100, gap=0 → cellSize=100, cols at 0, 100, 200
		r := layouts.Square([]float64{1, 1, 1}, 300, 100, 0)
		if len(r.Boxes) != 3 {
			t.Fatalf("expected 3 boxes, got %d", len(r.Boxes))
		}
		checkBox(t, r.Boxes[0], 0, 0, 100, 100)
		checkBox(t, r.Boxes[1], 0, 100, 100, 100)
		checkBox(t, r.Boxes[2], 0, 200, 100, 100)
		if !near(r.ContainerHeight, 100) {
			t.Errorf("containerHeight: got %v, want 100", r.ContainerHeight)
		}
	})

	t.Run("six_photos_two_rows_no_gap", func(t *testing.T) {
		// cW=300, tS=100, gap=0 → 3 columns; row 0 at top=0, row 1 at top=100
		r := layouts.Square(repeat(6, 1), 300, 100, 0)
		for i := 0; i < 3; i++ {
			if !near(r.Boxes[i].Top, 0) {
				t.Errorf("photo %d: top=%v, want 0 (row 0)", i, r.Boxes[i].Top)
			}
		}
		for i := 3; i < 6; i++ {
			if !near(r.Boxes[i].Top, 100) {
				t.Errorf("photo %d: top=%v, want 100 (row 1)", i, r.Boxes[i].Top)
			}
		}
		if !near(r.ContainerHeight, 200) {
			t.Errorf("containerHeight: got %v, want 200", r.ContainerHeight)
		}
	})

	t.Run("spread_increases_cell_size_to_fill_container", func(t *testing.T) {
		// cW=302, tS=100, gap=0 → perChunk=3, remaining=2, spread=ceil(2/3)=1, cellSize=101
		r := layouts.Square([]float64{1, 1, 1}, 302, 100, 0)
		checkBox(t, r.Boxes[0], 0, 0, 101, 101)
		checkBox(t, r.Boxes[1], 0, 101, 101, 101)
		checkBox(t, r.Boxes[2], 0, 202, 101, 101)
		if !near(r.ContainerHeight, 101) {
			t.Errorf("containerHeight: got %v, want 101", r.ContainerHeight)
		}
	})

	t.Run("gap_offsets_column_left_positions", func(t *testing.T) {
		// cW=306, tS=100, gap=3 → perChunk=3, remaining=0, cellSize=100
		// col lefts: 0*(100+3)=0, 1*103=103, 2*103=206
		r := layouts.Square([]float64{1, 1, 1}, 306, 100, 3)
		checkBox(t, r.Boxes[0], 0, 0, 100, 100)
		checkBox(t, r.Boxes[1], 0, 103, 100, 100)
		checkBox(t, r.Boxes[2], 0, 206, 100, 100)
	})

	t.Run("gap_offsets_row_top_positions", func(t *testing.T) {
		// cW=306, tS=100, gap=3 → cellSize=100
		// After row 0: col heights = 100+3=103, so row 1 starts at top=103.
		r := layouts.Square(repeat(6, 1), 306, 100, 3)
		for i := 0; i < 3; i++ {
			if !near(r.Boxes[i].Top, 0) {
				t.Errorf("photo %d: top=%v, want 0 (row 0)", i, r.Boxes[i].Top)
			}
		}
		for i := 3; i < 6; i++ {
			if !near(r.Boxes[i].Top, 103) {
				t.Errorf("photo %d: top=%v, want 103 (row 1)", i, r.Boxes[i].Top)
			}
		}
		if !near(r.ContainerHeight, 206) {
			t.Errorf("containerHeight: got %v, want 206", r.ContainerHeight)
		}
	})

	t.Run("single_column_when_container_narrower_than_target", func(t *testing.T) {
		// cW=50, tS=100, gap=0 → perChunk=max(1,0)=1, spread=-50, cellSize=50
		r := layouts.Square([]float64{1, 1}, 50, 100, 0)
		if len(r.Boxes) != 2 {
			t.Fatalf("expected 2 boxes, got %d", len(r.Boxes))
		}
		checkBox(t, r.Boxes[0], 0, 0, 50, 50)
		checkBox(t, r.Boxes[1], 50, 0, 50, 50)
		if !near(r.ContainerHeight, 100) {
			t.Errorf("containerHeight: got %v, want 100", r.ContainerHeight)
		}
	})
}

// ════════════════════════════════════════════════════════════════════════════
// Grid
// ════════════════════════════════════════════════════════════════════════════

func TestGrid(t *testing.T) {
	t.Run("empty_input_returns_default", func(t *testing.T) {
		r := layouts.Grid(nil, 300, 100, 0)
		if len(r.Boxes) != 0 {
			t.Fatalf("expected 0 boxes, got %d", len(r.Boxes))
		}
		if r.ContainerHeight != 0 {
			t.Errorf("containerHeight: got %v, want 0", r.ContainerHeight)
		}
	})

	t.Run("zero_container_width_returns_default", func(t *testing.T) {
		r := layouts.Grid([]float64{1}, 0, 100, 0)
		if len(r.Boxes) != 0 {
			t.Fatalf("expected 0 boxes, got %d", len(r.Boxes))
		}
	})

	t.Run("photo_height_is_col_width_divided_by_ratio_floored", func(t *testing.T) {
		// colWidth=100; h = floor(100 / ratio)
		r := layouts.Grid([]float64{2, 1, 0.5}, 300, 100, 0)
		if len(r.Boxes) != 3 {
			t.Fatalf("expected 3 boxes, got %d", len(r.Boxes))
		}
		checkBox(t, r.Boxes[0], 0, 0, 100, 50)    // floor(100/2.0) = 50
		checkBox(t, r.Boxes[1], 0, 100, 100, 100)  // floor(100/1.0) = 100
		checkBox(t, r.Boxes[2], 0, 200, 100, 200)  // floor(100/0.5) = 200
	})

	t.Run("invalid_ratio_treated_as_one", func(t *testing.T) {
		// ratio=0 or negative → falls back to 1 → h=floor(100/1)=100
		r := layouts.Grid([]float64{0, -2}, 200, 100, 0)
		if !near(r.Boxes[0].Height, 100) {
			t.Errorf("box 0 height: got %v, want 100", r.Boxes[0].Height)
		}
		if !near(r.Boxes[1].Height, 100) {
			t.Errorf("box 1 height: got %v, want 100", r.Boxes[1].Height)
		}
	})

	t.Run("all_photos_in_same_row_share_top", func(t *testing.T) {
		// 3 photos into 3 columns → all in row 0, all at top=0
		r := layouts.Grid([]float64{2, 1, 0.5}, 300, 100, 0)
		for i, b := range r.Boxes {
			if !near(b.Top, 0) {
				t.Errorf("photo %d: top=%v, want 0", i, b.Top)
			}
		}
	})

	t.Run("second_row_starts_at_tallest_column_of_first_row", func(t *testing.T) {
		// Row 0: heights 50, 100, 200 → max=200 → row 1 syncs to top=200
		ratios := []float64{2, 1, 0.5, 1, 1, 1}
		r := layouts.Grid(ratios, 300, 100, 0)
		for i := 3; i < 6; i++ {
			if !near(r.Boxes[i].Top, 200) {
				t.Errorf("photo %d: top=%v, want 200 (row 1)", i, r.Boxes[i].Top)
			}
		}
	})

	t.Run("second_row_correct_heights", func(t *testing.T) {
		// Row 1: all ratio=1 → h=floor(100/1)=100; containerHeight=200(sync)+100=300
		ratios := []float64{2, 1, 0.5, 1, 1, 1}
		r := layouts.Grid(ratios, 300, 100, 0)
		checkBox(t, r.Boxes[3], 200, 0, 100, 100)
		checkBox(t, r.Boxes[4], 200, 100, 100, 100)
		checkBox(t, r.Boxes[5], 200, 200, 100, 100)
		if !near(r.ContainerHeight, 300) {
			t.Errorf("containerHeight: got %v, want 300", r.ContainerHeight)
		}
	})

	t.Run("gap_offsets_row_and_column_positions", func(t *testing.T) {
		// cW=306, tW=100, gap=3 → perChunk=3, colWidth=100
		// col lefts: 0, 103, 206; after row 0: col heights=100+3=103, row 1 at top=103
		r := layouts.Grid(repeat(6, 1), 306, 100, 3)
		checkBox(t, r.Boxes[0], 0, 0, 100, 100)
		checkBox(t, r.Boxes[1], 0, 103, 100, 100)
		checkBox(t, r.Boxes[2], 0, 206, 100, 100)
		for i := 3; i < 6; i++ {
			if !near(r.Boxes[i].Top, 103) {
				t.Errorf("photo %d: top=%v, want 103 (row 1)", i, r.Boxes[i].Top)
			}
		}
		if !near(r.ContainerHeight, 206) {
			t.Errorf("containerHeight: got %v, want 206", r.ContainerHeight)
		}
	})
}

// ════════════════════════════════════════════════════════════════════════════
// Masonry
// ════════════════════════════════════════════════════════════════════════════

func TestMasonry(t *testing.T) {
	t.Run("empty_input_returns_default", func(t *testing.T) {
		r := layouts.Masonry(nil, 300, 100, 0)
		if len(r.Boxes) != 0 {
			t.Fatalf("expected 0 boxes, got %d", len(r.Boxes))
		}
		if r.ContainerHeight != 0 {
			t.Errorf("containerHeight: got %v, want 0", r.ContainerHeight)
		}
	})

	t.Run("zero_container_width_returns_default", func(t *testing.T) {
		r := layouts.Masonry([]float64{1}, 0, 100, 0)
		if len(r.Boxes) != 0 {
			t.Fatalf("expected 0 boxes, got %d", len(r.Boxes))
		}
	})

	t.Run("equal_ratios_fill_columns_left_to_right", func(t *testing.T) {
		// All columns start at 0; ties go to the lowest-index column (strict <).
		r := layouts.Masonry([]float64{1, 1, 1}, 300, 100, 0)
		if len(r.Boxes) != 3 {
			t.Fatalf("expected 3 boxes, got %d", len(r.Boxes))
		}
		checkBox(t, r.Boxes[0], 0, 0, 100, 100)
		checkBox(t, r.Boxes[1], 0, 100, 100, 100)
		checkBox(t, r.Boxes[2], 0, 200, 100, 100)
		if !near(r.ContainerHeight, 100) {
			t.Errorf("containerHeight: got %v, want 100", r.ContainerHeight)
		}
	})

	t.Run("portrait_photo_delays_its_column", func(t *testing.T) {
		// Photo 0 (ratio=0.5) → col0 h=200; photos 1,2 → col1,col2 h=100; photo 3 → col1 (tie)
		r := layouts.Masonry([]float64{0.5, 1, 1, 1}, 300, 100, 0)
		checkBox(t, r.Boxes[0], 0, 0, 100, 200)
		checkBox(t, r.Boxes[1], 0, 100, 100, 100)
		checkBox(t, r.Boxes[2], 0, 200, 100, 100)
		checkBox(t, r.Boxes[3], 100, 100, 100, 100) // stacked on top of photo 1
		if !near(r.ContainerHeight, 200) {
			t.Errorf("containerHeight: got %v, want 200", r.ContainerHeight)
		}
	})

	t.Run("landscape_photo_produces_short_cell", func(t *testing.T) {
		// h = colWidth / ratio = 100 / 2 = 50
		r := layouts.Masonry([]float64{2, 1, 1}, 300, 100, 0)
		checkBox(t, r.Boxes[0], 0, 0, 100, 50) // wide → short cell
		checkBox(t, r.Boxes[1], 0, 100, 100, 100)
		checkBox(t, r.Boxes[2], 0, 200, 100, 100)
		if !near(r.ContainerHeight, 100) {
			t.Errorf("containerHeight: got %v, want 100", r.ContainerHeight)
		}
	})

	t.Run("next_photo_goes_to_shortest_column", func(t *testing.T) {
		// Portrait in col0 (h=200) keeps it out of rotation until others grow.
		// Photo 3 (ratio=0.5) goes to col1 (tied with col2 at 100, first wins).
		r := layouts.Masonry([]float64{0.5, 1, 1, 0.5}, 300, 100, 0)
		if !near(r.Boxes[3].Left, 100) {
			t.Errorf("photo 3 left: got %v, want 100 (col1)", r.Boxes[3].Left)
		}
		if !near(r.Boxes[3].Top, 100) {
			t.Errorf("photo 3 top: got %v, want 100", r.Boxes[3].Top)
		}
		if !near(r.Boxes[3].Height, 200) {
			t.Errorf("photo 3 height: got %v, want 200", r.Boxes[3].Height)
		}
		if !near(r.ContainerHeight, 300) {
			t.Errorf("containerHeight: got %v, want 300", r.ContainerHeight)
		}
	})

	t.Run("invalid_ratio_treated_as_one", func(t *testing.T) {
		r := layouts.Masonry([]float64{0, -1}, 200, 100, 0)
		if !near(r.Boxes[0].Height, 100) {
			t.Errorf("box 0 height: got %v, want 100", r.Boxes[0].Height)
		}
		if !near(r.Boxes[1].Height, 100) {
			t.Errorf("box 1 height: got %v, want 100", r.Boxes[1].Height)
		}
	})

	t.Run("gap_increases_column_heights", func(t *testing.T) {
		// cW=306, tW=100, gap=3 → colWidth=100; each photo adds h+gap=103 to its column
		r := layouts.Masonry([]float64{1, 1, 1}, 306, 100, 3)
		if !near(r.ContainerHeight, 103) {
			t.Errorf("containerHeight: got %v, want 103", r.ContainerHeight)
		}
	})

	t.Run("container_height_is_tallest_column", func(t *testing.T) {
		// ratios=[1, 0.5, 1]: col0=100, col1=200, col2=100 → max=200
		r := layouts.Masonry([]float64{1, 0.5, 1}, 300, 100, 0)
		if !near(r.ContainerHeight, 200) {
			t.Errorf("containerHeight: got %v, want 200", r.ContainerHeight)
		}
	})
}

// ════════════════════════════════════════════════════════════════════════════
// Justified
// ════════════════════════════════════════════════════════════════════════════

func TestJustified(t *testing.T) {
	t.Run("empty_input_returns_default", func(t *testing.T) {
		r := layouts.Justified(nil, 300, 100, 0)
		if len(r.Boxes) != 0 {
			t.Fatalf("expected 0 boxes, got %d", len(r.Boxes))
		}
		if r.ContainerHeight != 0 {
			t.Errorf("containerHeight: got %v, want 0", r.ContainerHeight)
		}
	})

	t.Run("zero_container_width_returns_default", func(t *testing.T) {
		r := layouts.Justified([]float64{1}, 0, 100, 0)
		if len(r.Boxes) != 0 {
			t.Fatalf("expected 0 boxes, got %d", len(r.Boxes))
		}
	})

	t.Run("non_positive_row_height_defaults_to_cw_over_four", func(t *testing.T) {
		// tRH=0 → effective=400/4=100; single photo (ratio=1): ideal=400>100, last-row default h=100
		r := layouts.Justified([]float64{1}, 400, 0, 0)
		if len(r.Boxes) != 1 {
			t.Fatalf("expected 1 box, got %d", len(r.Boxes))
		}
		if !near(r.Boxes[0].Height, 100) {
			t.Errorf("height: got %v, want 100", r.Boxes[0].Height)
		}
	})

	t.Run("single_landscape_photo_fills_full_row_width", func(t *testing.T) {
		// ratio=2, cW=200, tRH=100: ideal=100=target → h=100, w=2×100=200
		r := layouts.Justified([]float64{2}, 200, 100, 0)
		checkBox(t, r.Boxes[0], 0, 0, 200, 100)
		if !near(r.ContainerHeight, 100) {
			t.Errorf("containerHeight: got %v, want 100", r.ContainerHeight)
		}
	})

	t.Run("single_portrait_photo_placed_alone", func(t *testing.T) {
		// ratio=0.5, cW=200, tRH=100: ideal=400>target, last-row default h=100, w=0.5×100=50
		r := layouts.Justified([]float64{0.5}, 200, 100, 0)
		if !near(r.Boxes[0].Height, 100) {
			t.Errorf("height: got %v, want 100", r.Boxes[0].Height)
		}
		if !near(r.Boxes[0].Width, 50) {
			t.Errorf("width: got %v, want 50", r.Boxes[0].Width)
		}
	})

	t.Run("two_equal_photos_fill_row_exactly", func(t *testing.T) {
		// ratios=[1,1], cW=200, tRH=100: j=1 ideal=100 → h=100, each w=100
		r := layouts.Justified([]float64{1, 1}, 200, 100, 0)
		checkBox(t, r.Boxes[0], 0, 0, 100, 100)
		checkBox(t, r.Boxes[1], 0, 100, 100, 100)
		if !near(r.ContainerHeight, 100) {
			t.Errorf("containerHeight: got %v, want 100", r.ContainerHeight)
		}
	})

	t.Run("last_row_uses_target_height_without_stretching", func(t *testing.T) {
		// 4 photos, cW=300, tRH=100: row0=3 photos h=100; row1=1 photo at h=100 (target, not stretched)
		r := layouts.Justified(repeat(4, 1), 300, 100, 0)
		for i := 0; i < 3; i++ {
			if !near(r.Boxes[i].Top, 0) {
				t.Errorf("photo %d: top=%v, want 0 (row 0)", i, r.Boxes[i].Top)
			}
		}
		checkBox(t, r.Boxes[3], 100, 0, 100, 100)
		if !near(r.ContainerHeight, 200) {
			t.Errorf("containerHeight: got %v, want 200", r.ContainerHeight)
		}
	})

	t.Run("prefers_n_minus_one_photos_when_closer_to_target", func(t *testing.T) {
		// ratios=[35,35,50], cW=8400, tRH=100:
		//   after j=1: ideal=120 (|120-100|=20)
		//   after j=2: ideal=70  (|70-100|=30)  → prefer j=1, row_height=120
		r := layouts.Justified([]float64{35, 35, 50}, 8400, 100, 0)
		if len(r.Boxes) != 3 {
			t.Fatalf("expected 3 boxes, got %d", len(r.Boxes))
		}
		checkBox(t, r.Boxes[0], 0, 0, 4200, 120)
		checkBox(t, r.Boxes[1], 0, 4200, 4200, 120)
		checkBox(t, r.Boxes[2], 120, 0, 5000, 100) // last row at target height
		if !near(r.ContainerHeight, 220) {
			t.Errorf("containerHeight: got %v, want 220", r.ContainerHeight)
		}
	})

	t.Run("prefers_n_photos_when_closer_to_target", func(t *testing.T) {
		// ratios=[1,1,1], cW=300, tRH=100: j=2 ideal=100=target → use 3 photos
		r := layouts.Justified([]float64{1, 1, 1}, 300, 100, 0)
		for _, b := range r.Boxes {
			if !near(b.Height, 100) {
				t.Errorf("height: got %v, want 100", b.Height)
			}
		}
		if !near(r.ContainerHeight, 100) {
			t.Errorf("containerHeight: got %v, want 100", r.ContainerHeight)
		}
	})

	t.Run("spacing_is_subtracted_from_available_width", func(t *testing.T) {
		// ratios=[1,1], cW=202, tRH=100, spacing=2:
		//   j=1: avail=202-2=200, ideal=100 → h=100; photo 1 left=100+2=102
		r := layouts.Justified([]float64{1, 1}, 202, 100, 2)
		checkBox(t, r.Boxes[0], 0, 0, 100, 100)
		checkBox(t, r.Boxes[1], 0, 102, 100, 100)
		if !near(r.ContainerHeight, 100) {
			t.Errorf("containerHeight: got %v, want 100", r.ContainerHeight)
		}
	})

	t.Run("spacing_advances_current_top_between_rows", func(t *testing.T) {
		// 4 photos, cW=200, tRH=100, spacing=10:
		//   j=1: avail=190, ideal=95 → row_height=95; current_top = 95+10 = 105
		//   row 1 photos must start at top=105
		r := layouts.Justified(repeat(4, 1), 200, 100, 10)
		row0H := r.Boxes[0].Height
		row1Top := r.Boxes[0].Top + row0H + 10
		if !near(r.Boxes[2].Top, row1Top) {
			t.Errorf("photo 2 top: got %v, want %v", r.Boxes[2].Top, row1Top)
		}
		if !near(r.Boxes[3].Top, row1Top) {
			t.Errorf("photo 3 top: got %v, want %v", r.Boxes[3].Top, row1Top)
		}
	})

	t.Run("invalid_ratios_fall_back_to_one", func(t *testing.T) {
		// ratios=[0,-1] treated as [1,1] → same result as two_equal_photos_fill_row_exactly
		r := layouts.Justified([]float64{0, -1}, 200, 100, 0)
		if !near(r.Boxes[0].Height, 100) {
			t.Errorf("box 0 height: got %v, want 100", r.Boxes[0].Height)
		}
		if !near(r.Boxes[1].Height, 100) {
			t.Errorf("box 1 height: got %v, want 100", r.Boxes[1].Height)
		}
		if !near(r.ContainerHeight, 100) {
			t.Errorf("containerHeight: got %v, want 100", r.ContainerHeight)
		}
	})
}
