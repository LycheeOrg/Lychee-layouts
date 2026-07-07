use wasm_bindgen::prelude::*;

mod grid;
mod justified;
mod masonry;
mod square;
mod types;

fn to_js(result: types::LayoutResult) -> JsValue {
    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// justified(ratios, containerWidth, targetRowHeight, spacing) → {containerHeight, boxes}
///
/// Lays out photos in rows where every photo shares the same height and all
/// photos together fill the container width. Greedy Flickr algorithm.
#[wasm_bindgen]
pub fn justified(
    ratios: &[f64],
    container_width: f64,
    target_row_height: f64,
    spacing: f64,
) -> JsValue {
    to_js(justified::compute(ratios, container_width, target_row_height, spacing))
}

/// square(ratios, containerWidth, targetSize, gap) → {containerHeight, boxes}
///
/// Places all photos as uniform squares. Aspect ratios are ignored.
#[wasm_bindgen]
pub fn square(ratios: &[f64], container_width: f64, target_size: f64, gap: f64) -> JsValue {
    to_js(square::compute(ratios, container_width, target_size, gap))
}

/// masonry(ratios, containerWidth, targetWidth, gap) → {containerHeight, boxes}
///
/// Places each photo into the shortest column (Pinterest style).
#[wasm_bindgen]
pub fn masonry(ratios: &[f64], container_width: f64, target_width: f64, gap: f64) -> JsValue {
    to_js(masonry::compute(ratios, container_width, target_width, gap))
}

/// grid(ratios, containerWidth, targetWidth, gap) → {containerHeight, boxes}
///
/// Fixed columns with aligned row starts and variable photo heights.
#[wasm_bindgen]
pub fn grid(ratios: &[f64], container_width: f64, target_width: f64, gap: f64) -> JsValue {
    to_js(grid::compute(ratios, container_width, target_width, gap))
}
