use crate::types::{LayoutResult, PhotoBox};

struct Col {
    height: f64,
    left: f64,
}

/// Places all photos as uniform squares, filling rows left to right.
/// The aspect ratio of each photo is ignored — every cell is target_size × target_size
/// (spread evenly to fill the container width).
pub fn compute(ratios: &[f64], container_width: f64, target_size: f64, gap: f64) -> LayoutResult {
    if ratios.is_empty() || container_width <= 0.0 || target_size <= 0.0 {
        return LayoutResult::default();
    }

    let per_chunk = ((container_width + gap) / target_size).floor().max(1.0) as usize;
    let remaining =
        container_width - per_chunk as f64 * target_size - (per_chunk as f64 - 1.0) * gap;
    let spread = (remaining / per_chunk as f64).ceil();
    let cell_size = target_size + spread;

    let mut cols: Vec<Col> = (0..per_chunk)
        .map(|i| Col {
            height: 0.0,
            left: i as f64 * (cell_size + gap),
        })
        .collect();

    let mut boxes = vec![PhotoBox::default(); ratios.len()];
    let mut idx = 0usize;

    for i in 0..ratios.len() {
        if idx == 0 {
            // Sync all columns to the tallest at the start of each row.
            let max_h = cols.iter().map(|c| c.height).fold(f64::NEG_INFINITY, |a, b| a.max(b));
            for col in &mut cols {
                col.height = max_h;
            }
        }

        let top = cols[idx].height;
        let left = cols[idx].left;
        boxes[i] = PhotoBox { top, left, width: cell_size, height: cell_size };
        cols[idx].height = top + cell_size + gap;
        idx = (idx + 1) % per_chunk;
    }

    let container_height =
        cols.iter().map(|c| c.height).fold(f64::NEG_INFINITY, |a, b| a.max(b));

    LayoutResult { container_height, boxes }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Assert a box's geometry with a human-readable error message.
    fn assert_box(b: &PhotoBox, top: f64, left: f64, width: f64, height: f64) {
        let eps = 1e-9;
        assert!((b.top    - top   ).abs() < eps, "top:    {} ≠ {}", b.top,    top);
        assert!((b.left   - left  ).abs() < eps, "left:   {} ≠ {}", b.left,   left);
        assert!((b.width  - width ).abs() < eps, "width:  {} ≠ {}", b.width,  width);
        assert!((b.height - height).abs() < eps, "height: {} ≠ {}", b.height, height);
    }

    // ── guard clauses ────────────────────────────────────────────────────────

    #[test]
    fn empty_input_returns_default() {
        let r = compute(&[], 300.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 0);
        assert_eq!(r.container_height, 0.0);
    }

    #[test]
    fn zero_container_width_returns_default() {
        let r = compute(&[1.0, 1.0], 0.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 0);
    }

    #[test]
    fn zero_target_size_returns_default() {
        let r = compute(&[1.0, 1.0], 300.0, 0.0, 0.0);
        assert_eq!(r.boxes.len(), 0);
    }

    // ── cell geometry ────────────────────────────────────────────────────────

    #[test]
    fn all_cells_are_square_regardless_of_ratio() {
        // Ratios must not influence cell dimensions.
        let r = compute(&[2.0, 0.5, 3.0, 0.1, 1.0], 300.0, 100.0, 0.0);
        for b in &r.boxes {
            let eps = 1e-9;
            assert!((b.width - b.height).abs() < eps,
                "cell is not square: {}×{}", b.width, b.height);
        }
    }

    #[test]
    fn single_photo_no_gap() {
        // cW=300, tS=100, gap=0 → perChunk=3, spread=0, cellSize=100
        let r = compute(&[1.0], 300.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 1);
        assert_box(&r.boxes[0], 0.0, 0.0, 100.0, 100.0);
        assert_eq!(r.container_height, 100.0);
    }

    // ── column and row layout ────────────────────────────────────────────────

    #[test]
    fn three_photos_fill_one_row_no_gap() {
        // cW=300, tS=100, gap=0 → cellSize=100, cols at 0, 100, 200
        let r = compute(&[1.0, 1.0, 1.0], 300.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 3);
        assert_box(&r.boxes[0], 0.0,   0.0, 100.0, 100.0);
        assert_box(&r.boxes[1], 0.0, 100.0, 100.0, 100.0);
        assert_box(&r.boxes[2], 0.0, 200.0, 100.0, 100.0);
        assert_eq!(r.container_height, 100.0);
    }

    #[test]
    fn six_photos_two_rows_no_gap() {
        // cW=300, tS=100, gap=0 → 3 columns, row 1 at top=100
        let r = compute(&[1.0; 6], 300.0, 100.0, 0.0);
        for i in 0..3 {
            assert_eq!(r.boxes[i].top, 0.0,   "photo {i} should be in row 0");
        }
        for i in 3..6 {
            assert_eq!(r.boxes[i].top, 100.0, "photo {i} should be in row 1");
        }
        assert_eq!(r.container_height, 200.0);
    }

    #[test]
    fn spread_increases_cell_size_to_fill_container() {
        // cW=302, tS=100, gap=0 → perChunk=3, remaining=2, spread=ceil(2/3)=1, cellSize=101
        let r = compute(&[1.0, 1.0, 1.0], 302.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 3);
        assert_box(&r.boxes[0], 0.0,   0.0, 101.0, 101.0);
        assert_box(&r.boxes[1], 0.0, 101.0, 101.0, 101.0);
        assert_box(&r.boxes[2], 0.0, 202.0, 101.0, 101.0);
        assert_eq!(r.container_height, 101.0);
    }

    #[test]
    fn gap_offsets_column_left_positions() {
        // cW=306, tS=100, gap=3 → perChunk=3, remaining=0, cellSize=100
        // col lefts: 0, 103, 206
        let r = compute(&[1.0, 1.0, 1.0], 306.0, 100.0, 3.0);
        assert_box(&r.boxes[0], 0.0,   0.0, 100.0, 100.0);
        assert_box(&r.boxes[1], 0.0, 103.0, 100.0, 100.0);
        assert_box(&r.boxes[2], 0.0, 206.0, 100.0, 100.0);
    }

    #[test]
    fn gap_offsets_row_top_positions() {
        // cW=306, tS=100, gap=3 → cellSize=100, gap=3
        // After row 0 each column height = 100+3 = 103, so row 1 starts at top=103.
        let r = compute(&[1.0; 6], 306.0, 100.0, 3.0);
        for i in 0..3 {
            assert_eq!(r.boxes[i].top,   0.0, "photo {i} should be in row 0");
        }
        for i in 3..6 {
            assert_eq!(r.boxes[i].top, 103.0, "photo {i} should be in row 1");
        }
        assert_eq!(r.container_height, 206.0);
    }

    #[test]
    fn single_column_when_container_narrower_than_target() {
        // cW=50, tS=100, gap=0 → perChunk=max(1, floor(0.5))=1
        // remaining=50-100=-50, spread=-50, cellSize=50
        let r = compute(&[1.0, 1.0], 50.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 2);
        assert_box(&r.boxes[0],  0.0, 0.0, 50.0, 50.0);
        assert_box(&r.boxes[1], 50.0, 0.0, 50.0, 50.0);
        assert_eq!(r.container_height, 100.0);
    }
}
