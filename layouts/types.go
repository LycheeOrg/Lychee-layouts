package layouts

// Box represents the computed position and dimensions of a single photo.
type Box struct {
	Top    float64 `json:"top"`
	Left   float64 `json:"left"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// LayoutResult is the output of any layout algorithm.
type LayoutResult struct {
	ContainerHeight float64 `json:"containerHeight"`
	Boxes           []Box   `json:"boxes"`
}
