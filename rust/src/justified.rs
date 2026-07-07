use crate::types::{LayoutResult, PhotoBox};

/// Lays out photos in rows where every photo in a row shares the same height
/// and all photos together fill the container width exactly.
/// Greedy implementation of the Flickr justified-layout algorithm.
///
/// The last (potentially incomplete) row is set to target_row_height without
/// stretching, matching the npm package's default behaviour.
pub fn compute(
    ratios: &[f64],
    container_width: f64,
    target_row_height: f64,
    spacing: f64,
) -> LayoutResult {
    if ratios.is_empty() || container_width <= 0.0 {
        return LayoutResult::default();
    }

    let target_row_height = if target_row_height <= 0.0 {
        container_width / 4.0
    } else {
        target_row_height
    };

    let mut boxes = Vec::with_capacity(ratios.len());
    let mut current_top = 0.0f64;
    let mut i = 0usize;

    while i < ratios.len() {
        let start = i;

        // Defaults for the last partial row: all remaining photos at target_row_height.
        let mut row_end = ratios.len() - 1;
        let mut row_height = target_row_height;

        let mut sum_ratios = 0.0f64;

        'row: for j in start..ratios.len() {
            let r = if ratios[j] <= 0.0 { 1.0 } else { ratios[j] };
            sum_ratios += r;

            let n = (j - start + 1) as f64;
            // Guard against pathological cases where spacing alone exceeds container width.
            let avail_width = (container_width - (n - 1.0) * spacing).max(container_width * 0.1);
            let ideal_height = avail_width / sum_ratios;

            if ideal_height <= target_row_height {
                // Adding photo j brought the row height to or below target.
                if j == start {
                    // Single photo — no choice but to use it.
                    row_end = j;
                    row_height = ideal_height;
                } else {
                    // Compare j vs j-1 photos: pick the height closest to target.
                    let r_j = if ratios[j] <= 0.0 { 1.0 } else { ratios[j] };
                    let prev_sum = sum_ratios - r_j;
                    let prev_avail = container_width - (n - 2.0) * spacing;
                    let prev_height = prev_avail / prev_sum;

                    if (prev_height - target_row_height).abs()
                        < (ideal_height - target_row_height).abs()
                    {
                        row_end = j - 1;
                        row_height = prev_height;
                    } else {
                        row_end = j;
                        row_height = ideal_height;
                    }
                }
                break 'row;
            }
            // ideal_height > target_row_height: row is not full yet.
            // Keep going; row_end / row_height retain their "last partial row" defaults.
        }

        let mut left = 0.0f64;
        for k in start..=row_end {
            let r = if ratios[k] <= 0.0 { 1.0 } else { ratios[k] };
            let w = r * row_height;
            boxes.push(PhotoBox {
                top: current_top,
                left,
                width: w,
                height: row_height,
            });
            left += w + spacing;
        }

        current_top += row_height + spacing;
        i = row_end + 1;
    }

    let container_height = boxes.last().map(|b| b.top + b.height).unwrap_or(0.0);

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

    #[test]
    fn non_positive_row_height_defaults_to_container_width_over_four() {
        // target_row_height=0 → effective = cW/4 = 400/4 = 100
        // Single photo (ratio=1): ideal=400 > 100, loop exits → last-row default, h=100
        let r = compute(&[1.0], 400.0, 0.0, 0.0);
        assert_eq!(r.boxes.len(), 1);
        assert_eq!(r.boxes[0].height, 100.0);
    }

    // ── single-photo row ─────────────────────────────────────────────────────

    #[test]
    fn single_landscape_photo_fills_full_row_width() {
        // ratio=2, cW=200, tRH=100, spacing=0
        // ideal = 200 / 2 = 100 = target → h=100, w=2×100=200
        let r = compute(&[2.0], 200.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 1);
        assert_box(&r.boxes[0], 0.0, 0.0, 200.0, 100.0);
        assert_eq!(r.container_height, 100.0);
    }

    #[test]
    fn single_portrait_photo_placed_alone() {
        // ratio=0.5, cW=200, tRH=100, spacing=0
        // ideal = 200/0.5 = 400 > target → last-row default, h=100, w=0.5×100=50
        let r = compute(&[0.5], 200.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 1);
        assert_eq!(r.boxes[0].height, 100.0);
        assert_eq!(r.boxes[0].width,   50.0);
    }

    // ── row-fill logic ───────────────────────────────────────────────────────

    #[test]
    fn two_equal_photos_fill_row_exactly() {
        // ratios=[1,1], cW=200, tRH=100, spacing=0
        // j=1: ideal = 200/2 = 100 = target → h=100, each w=100
        let r = compute(&[1.0, 1.0], 200.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 2);
        assert_box(&r.boxes[0], 0.0,   0.0, 100.0, 100.0);
        assert_box(&r.boxes[1], 0.0, 100.0, 100.0, 100.0);
        assert_eq!(r.container_height, 100.0);
    }

    #[test]
    fn last_row_uses_target_height_without_stretching() {
        // ratios=[1,1,1,1], cW=300, tRH=100, spacing=0
        // Row 0: j=2, ideal=100=target → 3 photos, h=100
        // Row 1: 1 photo left; ideal=300 > target → last-row default, h=100
        let r = compute(&[1.0; 4], 300.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 4);
        // Row 0
        for i in 0..3 { assert_eq!(r.boxes[i].top, 0.0,   "photo {i} must be in row 0"); }
        // Row 1
        assert_box(&r.boxes[3], 100.0, 0.0, 100.0, 100.0);
        assert_eq!(r.container_height, 200.0);
    }

    #[test]
    fn prefers_n_minus_one_photos_when_closer_to_target() {
        // ratios=[35, 35, 50], cW=8400, tRH=100, spacing=0
        //   after j=1: sum=70,  ideal=120  (20 above target)
        //   after j=2: sum=120, ideal=70   (30 below target)
        //   |120-100|=20 < |70-100|=30 → stop at j=1, row_height=120
        let r = compute(&[35.0, 35.0, 50.0], 8400.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 3);
        // Row 0: two photos at height=120
        assert_box(&r.boxes[0],   0.0,    0.0, 4200.0, 120.0);
        assert_box(&r.boxes[1],   0.0, 4200.0, 4200.0, 120.0);
        // Row 1: one photo at target height=100
        assert_box(&r.boxes[2], 120.0,    0.0, 5000.0, 100.0);
        assert_eq!(r.container_height, 220.0);
    }

    #[test]
    fn prefers_n_photos_when_closer_to_target() {
        // Symmetric check: when adding the Nth photo gets closer to target, keep it.
        // ratios=[1,1,1], cW=300, tRH=100, spacing=0
        //   j=2: sum=3, ideal=100=target → |100-100|=0, pick j=2
        let r = compute(&[1.0, 1.0, 1.0], 300.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 3);
        for b in &r.boxes { assert_eq!(b.height, 100.0); }
        assert_eq!(r.container_height, 100.0);
    }

    // ── spacing ──────────────────────────────────────────────────────────────

    #[test]
    fn spacing_is_subtracted_from_available_width() {
        // ratios=[1,1], cW=202, tRH=100, spacing=2
        // j=1: avail = 202-2 = 200, ideal = 200/2 = 100 → h=100, w=100 each
        // left of photo 1 = 100 + 2 (spacing) = 102
        let r = compute(&[1.0, 1.0], 202.0, 100.0, 2.0);
        assert_eq!(r.boxes.len(), 2);
        assert_box(&r.boxes[0], 0.0,   0.0, 100.0, 100.0);
        assert_box(&r.boxes[1], 0.0, 102.0, 100.0, 100.0);
        assert_eq!(r.container_height, 100.0);
    }

    #[test]
    fn spacing_advances_current_top_between_rows() {
        // Two full rows of 2 photos each, with vertical spacing between them.
        // ratios=[1,1,1,1], cW=200, tRH=100, spacing=10
        // Row 0: j=1: avail=190, ideal=95 <= 100 → prefer j vs j-1 (95 closer)
        //   row_height=95, current_top → 95+10=105
        // Row 1: same row_height=95, top=105
        let r = compute(&[1.0; 4], 200.0, 100.0, 10.0);
        assert_eq!(r.boxes.len(), 4);
        // All row-0 photos must share the same top.
        assert_eq!(r.boxes[0].top, r.boxes[1].top);
        // Row-1 photos must start after row-0 height + spacing.
        let row0_h  = r.boxes[0].height;
        let row1_top = r.boxes[0].top + row0_h + 10.0;
        let eps = 1e-9;
        assert!((r.boxes[2].top - row1_top).abs() < eps);
        assert!((r.boxes[3].top - row1_top).abs() < eps);
    }

    // ── invalid input ────────────────────────────────────────────────────────

    #[test]
    fn invalid_ratios_fall_back_to_one() {
        // ratios=[0, -1] treated as [1, 1]; same result as two_equal_photos_fill_row
        let r = compute(&[0.0, -1.0], 200.0, 100.0, 0.0);
        assert_eq!(r.boxes.len(), 2);
        assert_eq!(r.boxes[0].height, 100.0);
        assert_eq!(r.boxes[1].height, 100.0);
        assert_eq!(r.container_height, 100.0);
    }
}
