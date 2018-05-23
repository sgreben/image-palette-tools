package palette

import "math"

func hrgb(v1, v2, h float64) float64 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	switch {
	case 6*h < 1:
		return (v1 + (v2-v1)*6*h)
	case 2*h < 1:
		return v2
	case 3*h < 2:
		return v1 + (v2-v1)*((2.0/3.0)-h)*6
	}
	return v1
}

func rgb(h, s, l float64) (r, g, b uint8) {
	if s == 0 {
		r = uint8(255 * l)
		g = uint8(255 * l)
		b = uint8(255 * l)
		return
	}

	var v1, v2 float64
	if l < 0.5 {
		v2 = l * (1 + s)
	} else {
		v2 = (l + s) - (s * l)
	}

	v1 = 2*l - v2

	r = uint8(255 * hrgb(v1, v2, h+(1.0/3.0)))
	g = uint8(255 * hrgb(v1, v2, h))
	b = uint8(255 * hrgb(v1, v2, h-(1.0/3.0)))

	return
}

func hsl(rb, gb, bb uint8) (h, s, l float64) {
	r := float64(rb) / 255.0
	g := float64(gb) / 255.0
	b := float64(bb) / 255.0

	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	l = (max + min) / 2
	delta := max - min
	if delta != 0 {
		if l < 0.5 {
			s = delta / (max + min)
		} else {
			s = delta / (2 - max - min)
		}
		r2 := (((max - r) / 6) + (delta / 2)) / delta
		g2 := (((max - g) / 6) + (delta / 2)) / delta
		b2 := (((max - b) / 6) + (delta / 2)) / delta
		switch {
		case r == max:
			h = b2 - g2
		case g == max:
			h = (1.0 / 3.0) + r2 - b2
		case b == max:
			h = (2.0 / 3.0) + g2 - r2
		}
	}

	switch {
	case h < 0:
		h += 1
	case h > 1:
		h -= 1
	}
	return
}
