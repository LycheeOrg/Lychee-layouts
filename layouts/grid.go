package layouts

import "math"

// Grid places photos in fixed columns where each row starts aligned but photos
// within a row can have different heights based on their aspect ratio.
// At the start of each new row every column is synced to the tallest column,
// creating a regular grid appearance despite variable photo heights.
//
// ratios:         aspect ratios (width/height) for each photo
// containerWidth: total width of the container in pixels
// targetWidth:    desired column width
// gap:            spacing between cells (horizontal and vertical)
func Grid(ratios []float64, containerWidth, targetWidth, gap float64) LayoutResult {
	if len(ratios) == 0 || containerWidth <= 0 || targetWidth <= 0 {
		return LayoutResult{Boxes: []Box{}}
	}

	perChunk := math.Floor((containerWidth + gap) / targetWidth)
	if perChunk < 1 {
		perChunk = 1
	}
	n := int(perChunk)

	remaining := containerWidth - perChunk*targetWidth - (perChunk-1)*gap
	spread := math.Ceil(remaining / perChunk)
	colWidth := targetWidth + spread

	type col struct {
		height float64
		left   float64
	}
	cols := make([]col, n)
	for i := range cols {
		cols[i].left = float64(i) * (colWidth + gap)
	}

	boxes := make([]Box, len(ratios))
	idx := 0

	for i, ratio := range ratios {
		if idx == 0 {
			// Sync all columns to the tallest at the start of each row.
			maxH := cols[0].height
			for _, c := range cols[1:] {
				if c.height > maxH {
					maxH = c.height
				}
			}
			for j := range cols {
				cols[j].height = maxH
			}
		}

		r := ratio
		if r <= 0 {
			r = 1
		}
		c := cols[idx]
		h := math.Floor(colWidth / r)
		boxes[i] = Box{
			Top:    c.height,
			Left:   c.left,
			Width:  colWidth,
			Height: h,
		}
		cols[idx].height = c.height + h + gap
		idx = (idx + 1) % n
	}

	maxH := cols[0].height
	for _, c := range cols[1:] {
		if c.height > maxH {
			maxH = c.height
		}
	}

	return LayoutResult{ContainerHeight: maxH, Boxes: boxes}
}
