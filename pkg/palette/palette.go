package palette

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/bugra/kmeans"
)

func imagePoints(cache *ColorCache, i image.Image) (out [][]float64) {
	size := i.Bounds().Size()
	out = make([][]float64, size.X*size.Y)
	j := 0
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			c := i.At(x, y)
			r, g, b, _ := c.RGBA()
			out[j] = cache.Get(r, g, b)
			j++
		}
	}
	return
}

func Render(palette []color.RGBA, size int) image.Image {
	p := make(color.Palette, len(palette))
	for i := range palette {
		c := palette[i]
		c.A = 255
		p[i] = c
	}
	i := image.NewPaletted(image.Rectangle{
		Max: image.Point{
			X: size * len(palette),
			Y: size,
		},
	}, p)
	for j := range palette {
		for x := j * size; x < (j+1)*size; x++ {
			for y := 0; y < size; y++ {
				i.SetColorIndex(x, y, uint8(j))
			}
		}
	}
	return i
}

// Cluster clusters palettes
func Cluster(cache *PaletteCache, k int, ps [][]color.RGBA) ([]int, [][]color.RGBA, error) {
	if len(ps) == 0 {
		return nil, nil, nil
	}
	n := len(ps[0])

	points := make([][]float64, len(ps))
	for i := range ps {
		points[i] = cache.Get(ps[i])
	}

	labels, err := kmeans.Kmeans(points, k, kmeans.EuclideanDistance, int(math.MaxInt32))
	if err != nil {
		return nil, nil, err
	}

	centroidPoints := make([]kmeans.Observation, k)
	centroidPointCount := make([]uint64, k)
	for i, label := range labels {
		if centroidPoints[label] == nil {
			centroidPoints[label] = make(kmeans.Observation, n*3)
		}
		centroidPoints[label].Add(kmeans.Observation(points[i]))
		centroidPointCount[label]++
	}

	centroid := make([][]color.RGBA, k)
	for j, point := range centroidPoints {
		count := float64(centroidPointCount[j])
		centroid[j] = make([]color.RGBA, n)
		for i := 0; i < n; i++ {
			h := point[0+3*i] / count / weightH
			s := point[1+3*i] / count / weightS
			l := point[2+3*i] / count / weightL
			r, g, b := rgb(h, s, l)
			centroid[j][i] = color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
		}
	}

	return labels, centroid, nil
}

// Extract extracts a `k`-color palette from an image
func Extract(cache *ColorCache, k int, i image.Image) ([]color.RGBA, error) {
	points := imagePoints(cache, i)
	labels, err := kmeans.Kmeans(points, k, kmeans.EuclideanDistance, int(math.MaxInt32))
	if err != nil {
		return nil, err
	}

	centroidPoints := make([]kmeans.Observation, k)
	for label := range centroidPoints {
		centroidPoints[label] = make(kmeans.Observation, 3)
	}
	centroidPointCount := make([]uint64, k)
	for j, label := range labels {
		centroidPoints[label].Add(kmeans.Observation(points[j]))
		centroidPointCount[label]++
	}

	centroid := make([]color.RGBA, k)
	for j, point := range centroidPoints {
		n := float64(centroidPointCount[j])
		h := point[0] / n
		s := point[1] / n
		l := point[2] / n
		r, g, b := rgb(h, s, l)
		// r := uint32(point[0] / n * 255.0)
		// g := uint32(point[1] / n * 255.0)
		// b := uint32(point[2] / n * 255.0)
		centroid[j] = color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
	}

	sort.Slice(centroid, func(i int, j int) bool { return LessLHS(centroid, i, j) })

	return centroid, nil
}
