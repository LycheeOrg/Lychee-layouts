use serde::Serialize;

/// Computed position and dimensions of a single photo.
#[derive(Serialize, Default, Clone)]
pub struct PhotoBox {
    pub top: f64,
    pub left: f64,
    pub width: f64,
    pub height: f64,
}

/// Output of any layout algorithm.
#[derive(Serialize, Default)]
pub struct LayoutResult {
    #[serde(rename = "containerHeight")]
    pub container_height: f64,
    pub boxes: Vec<PhotoBox>,
}
