use crate::types::{LayoutResult, PhotoBox};

struct Col {
    height: f64,
    left: f64,
}

/// Places each photo into the shortest column (Pinterest-style staggered layout).
/// Unlike grid, rows are not aligned across columns.
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

    for (i, &ratio) in ratios.iter().enumerate() {
        // Find the shortest column.
        let min_idx = cols
            .iter()
            .enumerate()
            .min_by(|(_, a), (_, b)| a.height.total_cmp(&b.height))
            .map(|(i, _)| i)
            .unwrap_or(0);

        let r = if ratio <= 0.0 { 1.0 } else { ratio };
        let top = cols[min_idx].height;
        let left = cols[min_idx].left;
        let h = col_width / r;

        boxes[i] = PhotoBox { top, left, width: col_width, height: h };
        cols[min_idx].height = top + h + gap;
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

    // ── column selection ─────────────────────────────────────────────────────

    #[test]
    fn equal_ratios_fill_columns_left_to_right() {
        // All columns start at 0; min_by is stable so ties go to the first column.
        // Photo 0 → col0, photo 1 → col1, photo 2 → col2.
        let r = compute(&[1.0, 1.0, 1.0], 300.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 3);
        assert_box(&r.boxes[0], 0.0,   0.0, 100.0, 100.0);
        assert_box(&r.boxes[1], 0.0, 100.0, 100.0, 100.0);
        assert_box(&r.boxes[2], 0.0, 200.0, 100.0, 100.0);
        assert_eq!(r.container_height, 100.0);
    }

    #[test]
    fn portrait_photo_delays_its_column() {
        // Photo 0 (ratio=0.5) → col0, height=200.
        // Photos 1 and 2 (ratio=1.0) go to col1 and col2 (both at 0).
        // Photo 3 (ratio=1.0) goes to whichever of col1/col2 is shorter; both at 100, so col1.
        let r = compute(&[0.5, 1.0, 1.0, 1.0], 300.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 4);
        assert_box(&r.boxes[0],   0.0,   0.0, 100.0, 200.0); // col0, tall portrait
        assert_box(&r.boxes[1],   0.0, 100.0, 100.0, 100.0); // col1
        assert_box(&r.boxes[2],   0.0, 200.0, 100.0, 100.0); // col2
        assert_box(&r.boxes[3], 100.0, 100.0, 100.0, 100.0); // col1 again (tie → first)
        // col heights: 200, 200, 100 → max = 200
        assert_eq!(r.container_height, 200.0);
    }

    #[test]
    fn landscape_photo_produces_short_cell() {
        // h = colWidth / ratio = 100 / 2.0 = 50 (shorter than a square)
        let r = compute(&[2.0, 1.0, 1.0], 300.0, 100.0, 0.0);
        // Photo 0 → col0 (all 0): h=50; col0=50
        // Photo 1 → col1 (col0=50, col1=0): h=100; col1=100
        // Photo 2 → col2 (col0=50, col1=100, col2=0): h=100; col2=100
        assert_box(&r.boxes[0], 0.0,   0.0, 100.0,  50.0);
        assert_box(&r.boxes[1], 0.0, 100.0, 100.0, 100.0);
        assert_box(&r.boxes[2], 0.0, 200.0, 100.0, 100.0);
        assert_eq!(r.container_height, 100.0);
    }

    #[test]
    fn next_photo_goes_to_shortest_column() {
        // After placing a portrait (tall) in col0, col0 is skipped for the next photos
        // until the other columns catch up.
        // ratios=[0.5, 1.0, 1.0, 0.5]
        // Photo 0 → col0 (all 0): h=200; col0=200
        // Photo 1 → col1 (col0=200, col1=0): h=100; col1=100
        // Photo 2 → col2 (col0=200, col1=100, col2=0): h=100; col2=100
        // Photo 3 → col1 or col2 (tied at 100, first=col1): h=200; col1=300
        let r = compute(&[0.5, 1.0, 1.0, 0.5], 300.0, 100.0, 0.0);
        assert_eq!(r.boxes[3].left, 100.0); // must go to col1
        assert_eq!(r.boxes[3].top,  100.0); // stacked on top of photo 1
        assert_eq!(r.boxes[3].height, 200.0);
        assert_eq!(r.container_height, 300.0); // col1: 300
    }

    #[test]
    fn invalid_ratio_treated_as_one() {
        let r = compute(&[0.0, -1.0], 200.0, 100.0, 0.0);
        // Both treated as 1.0 → h = 100 each
        assert_eq!(r.boxes[0].height, 100.0);
        assert_eq!(r.boxes[1].height, 100.0);
    }

    #[test]
    fn gap_increases_column_heights() {
        // cW=306, tW=100, gap=3 → colWidth=100; after each photo, col.height += h + gap
        // 3 equal photos → each col gets h=100, then +3 gap → col heights = 103
        let r = compute(&[1.0, 1.0, 1.0], 306.0, 100.0, 3.0);
        assert_eq!(r.container_height, 103.0);
    }

    #[test]
    fn container_height_is_tallest_column() {
        // ratios=[1.0, 0.5, 1.0], cW=300, tW=100, gap=0
        // Photo 0 → col0: h=100; col0=100
        // Photo 1 → col1: h=200; col1=200
        // Photo 2 → col2: h=100; col2=100
        // max = 200
        let r = compute(&[1.0, 0.5, 1.0], 300.0, 100.0, 0.0);
        assert_eq!(r.container_height, 200.0);
    }
}
