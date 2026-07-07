package layouts

import "math"

// Justified lays out photos in rows where every photo in a row shares the same
// height and all photos together fill the container width exactly.
// This is a greedy implementation of the Flickr justified-layout algorithm,
// replacing the justified-layout npm dependency.
//
// The last (potentially incomplete) row is set to targetRowHeight without
// stretching, matching the npm package's default behaviour.
//
// ratios:          aspect ratios (width/height) for each photo
// containerWidth:  total width of the container in pixels
// targetRowHeight: desired height for each row
// spacing:         gap between photos (horizontal and vertical)
func Justified(ratios []float64, containerWidth, targetRowHeight, spacing float64) LayoutResult {
	if len(ratios) == 0 || containerWidth <= 0 {
		return LayoutResult{Boxes: []Box{}}
	}
	if targetRowHeight <= 0 {
		targetRowHeight = containerWidth / 4
	}

	boxes := make([]Box, 0, len(ratios))
	currentTop := 0.0
	i := 0

	for i < len(ratios) {
		start := i

		// Defaults for the "last partial row": use all remaining photos at targetRowHeight.
		rowEnd := len(ratios) - 1
		rowHeight := targetRowHeight

		sumRatios := 0.0
		for j := start; j < len(ratios); j++ {
			r := ratios[j]
			if r <= 0 {
				r = 1
			}
			sumRatios += r

			n := j - start + 1
			availWidth := containerWidth - float64(n-1)*spacing
			if availWidth < containerWidth*0.1 {
				// Pathological: too many photos for one row, place them anyway.
				availWidth = containerWidth
			}
			idealHeight := availWidth / sumRatios

			if idealHeight <= targetRowHeight {
				// Adding photo j brought the row height to or below target.
				if n == 1 {
					// Single photo — no choice but to use it.
					rowEnd = j
					rowHeight = idealHeight
				} else {
					// Compare using j vs j-1 photos: pick the height closest to target.
					prevSum := sumRatios - r
					prevAvailWidth := containerWidth - float64(n-2)*spacing
					prevHeight := prevAvailWidth / prevSum

					if math.Abs(prevHeight-targetRowHeight) < math.Abs(idealHeight-targetRowHeight) {
						rowEnd = j - 1
						rowHeight = prevHeight
					} else {
						rowEnd = j
						rowHeight = idealHeight
					}
				}
				break
			}
			// idealHeight > targetRowHeight: row is not full yet.
			// Keep going; rowEnd/rowHeight retain their "last partial row" defaults.
		}

		left := 0.0
		for k := start; k <= rowEnd; k++ {
			r := ratios[k]
			if r <= 0 {
				r = 1
			}
			w := r * rowHeight
			boxes = append(boxes, Box{
				Top:    currentTop,
				Left:   left,
				Width:  w,
				Height: rowHeight,
			})
			left += w + spacing
		}

		currentTop += rowHeight + spacing
		i = rowEnd + 1
	}

	containerHeight := 0.0
	if len(boxes) > 0 {
		last := boxes[len(boxes)-1]
		containerHeight = last.Top + last.Height
	}

	return LayoutResult{ContainerHeight: containerHeight, Boxes: boxes}
}
