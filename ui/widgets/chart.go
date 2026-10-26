package widgets

import "openwrt-monitor-tui/ui/style"

// Chart draws a series as a braille curve of w cells, newest on the right. A
// missing or zero sample sits on the baseline. In ASCII mode it falls back to
// the eight-level sparkline, because braille is not available there.
func Chart(values []float64, w int) string {
	if w <= 0 {
		return ""
	}
	if style.ASCII() {
		return Spark(values, w)
	}
	const rows = 4
	dots := w * 2

	y := make([]int, dots) // row index, 0 is the top dot row
	for i := range y {
		y[i] = rows - 1 // baseline
	}

	if len(values) > 0 {
		dense := resample(values, len(values)*2)
		if len(dense) > dots {
			dense = dense[len(dense)-dots:]
		}
		max := 0.0
		for _, v := range dense {
			if v > max {
				max = v
			}
		}
		off := dots - len(dense)
		for i, v := range dense {
			lvl := rows - 1
			if max > 0 {
				lvl = rows - 1 - int(v/max*float64(rows-1)+0.5)
			}
			y[off+i] = lvl
		}
	}

	grid := make([][rows]bool, dots)
	for x := 0; x < dots; x++ {
		grid[x][y[x]] = true
		if x > 0 {
			// Fill the vertical gap so a steep segment still looks connected.
			a, b := y[x-1], y[x]
			if a > b {
				a, b = b, a
			}
			for yy := a; yy <= b; yy++ {
				grid[x][yy] = true
			}
		}
	}

	out := make([]rune, w)
	for cx := 0; cx < w; cx++ {
		var mask rune
		for row := 0; row < rows; row++ {
			if grid[cx*2][row] {
				mask |= brailleBit(row, 0)
			}
			if grid[cx*2+1][row] {
				mask |= brailleBit(row, 1)
			}
		}
		out[cx] = 0x2800 + mask
	}
	return string(out)
}

// Spark is the ASCII/no-braille fallback: one character per sample, scaled to
// the series maximum.
func Spark(values []float64, w int) string {
	levels := []rune(style.Glyph.Spark)
	out := make([]rune, w)
	for i := range out {
		out[i] = levels[0]
	}
	if len(values) == 0 || w <= 0 {
		return string(out)
	}
	max := 0.0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if len(values) > w {
		values = values[len(values)-w:]
	}
	off := w - len(values)
	for i, v := range values {
		lvl := 0
		if max > 0 {
			lvl = int(v/max*float64(len(levels)-1) + 0.5)
		}
		out[off+i] = levels[lvl]
	}
	return string(out)
}

func brailleBit(row, col int) rune {
	if col == 0 {
		return [...]rune{0x01, 0x02, 0x04, 0x40}[row]
	}
	return [...]rune{0x08, 0x10, 0x20, 0x80}[row]
}

// resample scales a series to n points with linear interpolation.
func resample(s []float64, n int) []float64 {
	out := make([]float64, n)
	if len(s) == 0 {
		return out
	}
	if len(s) == 1 || n == 1 {
		for i := range out {
			out[i] = s[len(s)-1]
		}
		return out
	}
	for i := 0; i < n; i++ {
		f := float64(i) * float64(len(s)-1) / float64(n-1)
		j := int(f)
		if j >= len(s)-1 {
			out[i] = s[len(s)-1]
			continue
		}
		frac := f - float64(j)
		out[i] = s[j]*(1-frac) + s[j+1]*frac
	}
	return out
}
