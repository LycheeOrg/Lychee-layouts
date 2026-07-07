package layouts

import "math"

// Masonry places each photo into the shortest column, producing a Pinterest-style
// staggered layout. Unlike Grid, rows are not aligned across columns.
//
// ratios:         aspect ratios (width/height) for each photo
// containerWidth: total width of the container in pixels
// targetWidth:    desired column width
// gap:            spacing between cells (horizontal and vertical)
func Masonry(ratios []float64, containerWidth, targetWidth, gap float64) LayoutResult {
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

	for i, ratio := range ratios {
		// Find the shortest column.
		minIdx := 0
		for j := 1; j < n; j++ {
			if cols[j].height < cols[minIdx].height {
				minIdx = j
			}
		}

		r := ratio
		if r <= 0 {
			r = 1
		}
		c := cols[minIdx]
		h := colWidth / r
		boxes[i] = Box{
			Top:    c.height,
			Left:   c.left,
			Width:  colWidth,
			Height: h,
		}
		cols[minIdx].height = c.height + h + gap
	}

	maxH := cols[0].height
	for _, c := range cols[1:] {
		if c.height > maxH {
			maxH = c.height
		}
	}

	return LayoutResult{ContainerHeight: maxH, Boxes: boxes}
}
