use crate::types::{LayoutResult, PhotoBox};

struct Col {
    height: f64,
    left: f64,
}

/// Places photos in fixed columns where rows start aligned but photos within
/// a row can have different heights based on their aspect ratio.
/// At the start of each new row every column is synced to the tallest column.
pub fn compute(ratios: &[f64], container_width: f64, target_width: f64, gap: f64) -> LayoutResult {
    if ratios.is_empty() || container_width <= 0.0 || target_width <= 0.0 {
        return LayoutResult::default();
    }

    let per_chunk = ((container_width + gap) / target_width).floor().max(1.0) as usize;
    let remaining =
        container_width - per_chunk as f64 * target_width - (per_chunk as f64 - 1.0) * gap;
    let spread = (remaining / per_chunk as f64).ceil();
    let col_width = target_width + spread;

    let mut cols: Vec<Col> = (0..per_chunk)
        .map(|i| Col {
            height: 0.0,
            left: i as f64 * (col_width + gap),
        })
        .collect();

    let mut boxes = vec![PhotoBox::default(); ratios.len()];
    let mut idx = 0usize;

    for (i, &ratio) in ratios.iter().enumerate() {
        if idx == 0 {
            // Sync all columns to the tallest at the start of each row.
            let max_h = cols.iter().map(|c| c.height).fold(f64::NEG_INFINITY, |a, b| a.max(b));
            for col in &mut cols {
                col.height = max_h;
            }
        }

        let r = if ratio <= 0.0 { 1.0 } else { ratio };
        let top = cols[idx].height;
        let left = cols[idx].left;
        let h = (col_width / r).floor();

        boxes[i] = PhotoBox { top, left, width: col_width, height: h };
        cols[idx].height = top + h + gap;
        idx = (idx + 1) % per_chunk;
    }

    let container_height =
        cols.iter().map(|c| c.height).fold(f64::NEG_INFINITY, |a, b| a.max(b));

    LayoutResult { container_height, boxes }
}

#[cfg(test)]
mod tests {
    use super::*;

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
        let r = compute(&[1.0], 0.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 0);
    }

    // ── photo heights follow aspect ratio ────────────────────────────────────

    #[test]
    fn photo_height_is_col_width_divided_by_ratio_floored() {
        // colWidth=100; h = floor(100 / ratio)
        let r = compute(&[2.0, 1.0, 0.5], 300.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 3);
        assert_box(&r.boxes[0], 0.0,   0.0, 100.0,  50.0); // floor(100/2.0)  = 50
        assert_box(&r.boxes[1], 0.0, 100.0, 100.0, 100.0); // floor(100/1.0)  = 100
        assert_box(&r.boxes[2], 0.0, 200.0, 100.0, 200.0); // floor(100/0.5)  = 200
    }

    #[test]
    fn invalid_ratio_treated_as_one() {
        // ratio=0 or negative falls back to 1 → h = floor(100/1) = 100
        let r = compute(&[0.0, -2.0], 200.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 2);
        assert_eq!(r.boxes[0].height, 100.0);
        assert_eq!(r.boxes[1].height, 100.0);
    }

    // ── row alignment ─────────────────────────────────────────────────────────

    #[test]
    fn all_photos_in_same_row_share_top() {
        // 3 columns, 3 photos → all in one row, all at top=0
        let r = compute(&[2.0, 1.0, 0.5], 300.0, 100.0, 0.0);
        for b in &r.boxes {
            assert_eq!(b.top, 0.0, "all photos in row 0 must have top=0");
        }
    }

    #[test]
    fn second_row_starts_at_tallest_column_of_first_row() {
        // Row 0: heights 50, 100, 200 → max = 200
        // Row 1 photos must all start at top = 200
        let ratios = [2.0, 1.0, 0.5,   // row 0
                      1.0, 1.0, 1.0];  // row 1
        let r = compute(&ratios, 300.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 6);
        for i in 3..6 {
            assert_eq!(r.boxes[i].top, 200.0, "photo {i} must start at row-1 top=200");
        }
    }

    #[test]
    fn second_row_correct_heights() {
        // Row 1: all ratio=1 → h=100 each
        let ratios = [2.0, 1.0, 0.5, 1.0, 1.0, 1.0];
        let r = compute(&ratios, 300.0, 100.0, 0.0);
        assert_box(&r.boxes[3], 200.0,   0.0, 100.0, 100.0);
        assert_box(&r.boxes[4], 200.0, 100.0, 100.0, 100.0);
        assert_box(&r.boxes[5], 200.0, 200.0, 100.0, 100.0);
        // containerHeight = 200 (row-1 sync) + 100 (tallest in row 1) + 0 (no gap) = 300
        assert_eq!(r.container_height, 300.0);
    }

    #[test]
    fn gap_offsets_row_and_column_positions() {
        // cW=306, tW=100, gap=3 → perChunk=3, remaining=0, colWidth=100
        // col lefts: 0, 103, 206; after row 0: col heights = 100+3 = 103
        let r = compute(&[1.0; 6], 306.0, 100.0, 3.0);
        // Row 0
        assert_box(&r.boxes[0], 0.0,   0.0, 100.0, 100.0);
        assert_box(&r.boxes[1], 0.0, 103.0, 100.0, 100.0);
        assert_box(&r.boxes[2], 0.0, 206.0, 100.0, 100.0);
        // Row 1: sync to 103, all start at top=103
        for i in 3..6 {
            assert_eq!(r.boxes[i].top, 103.0, "photo {i} must be in row 1 at top=103");
        }
        assert_eq!(r.container_height, 206.0);
    }
}
