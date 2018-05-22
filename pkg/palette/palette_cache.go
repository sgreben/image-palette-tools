package palette

import (
	"encoding/binary"
	"image/color"
	"sync"

	"github.com/oxtoacart/bpool"
)

type PaletteCache struct {
	sync.Mutex
	Palettes   map[string][]float64
	BufferPool *bpool.BufferPool
}

func (c *PaletteCache) Key(p []color.RGBA) (key string) {
	buf := c.BufferPool.Get()
	defer c.BufferPool.Put(buf)
	for _, c := range p {
		binary.Write(buf, binary.LittleEndian, c.R)
		binary.Write(buf, binary.LittleEndian, c.G)
		binary.Write(buf, binary.LittleEndian, c.B)
	}
	key = buf.String()
	return
}

func (c *PaletteCache) Get(p []color.RGBA) []float64 {
	key := c.Key(p)
	c.Lock()
	defer c.Unlock()
	if point, ok := c.Palettes[key]; ok {
		return point
	}
	point := make([]float64, 3*len(p))
	for i, c := range p {
		point[0+3*i] = float64(c.R) / float64(0xff)
		point[1+3*i] = float64(c.G) / float64(0xff)
		point[2+3*i] = float64(c.B) / float64(0xff)
	}
	c.Palettes[key] = point
	return point
}

func NewPaletteCache(size int) *PaletteCache {
	return &PaletteCache{
		Palettes:   make(map[string][]float64, size),
		BufferPool: bpool.NewBufferPool(size),
	}
}
