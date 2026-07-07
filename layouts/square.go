package layouts

import "math"

// Square places all photos as uniform squares, filling rows left to right.
// The aspect ratio of each photo is ignored — every cell is targetSize × targetSize
// (spread evenly to fill the container width).
//
// ratios:         aspect ratios (width/height) for each photo — used only for count
// containerWidth: total width of the container in pixels
// targetSize:     desired width/height of each square cell
// gap:            spacing between cells (horizontal and vertical)
func Square(ratios []float64, containerWidth, targetSize, gap float64) LayoutResult {
	if len(ratios) == 0 || containerWidth <= 0 || targetSize <= 0 {
		return LayoutResult{Boxes: []Box{}}
	}

	perChunk := math.Floor((containerWidth + gap) / targetSize)
	if perChunk < 1 {
		perChunk = 1
	}
	n := int(perChunk)

	remaining := containerWidth - perChunk*targetSize - (perChunk-1)*gap
	spread := math.Ceil(remaining / perChunk)
	cellSize := targetSize + spread

	type col struct {
		height float64
		left   float64
	}
	cols := make([]col, n)
	for i := range cols {
		cols[i].left = float64(i) * (cellSize + gap)
	}

	boxes := make([]Box, len(ratios))
	idx := 0

	for i := range ratios {
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

		c := cols[idx]
		boxes[i] = Box{
			Top:    c.height,
			Left:   c.left,
			Width:  cellSize,
			Height: cellSize,
		}
		cols[idx].height = c.height + cellSize + gap
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
