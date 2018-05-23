package palette

import (
	"encoding/binary"
	"sync"

	"github.com/oxtoacart/bpool"
)

type ColorCache struct {
	sync.Mutex
	Colors     map[string][]float64
	BufferPool *bpool.BufferPool
}

func (c *ColorCache) Key(r, g, b uint32) (key string) {
	buf := c.BufferPool.Get()
	defer c.BufferPool.Put(buf)
	binary.Write(buf, binary.LittleEndian, r)
	binary.Write(buf, binary.LittleEndian, g)
	binary.Write(buf, binary.LittleEndian, b)
	key = buf.String()
	return
}

func (c *ColorCache) Get(r, g, b uint32) []float64 {
	key := c.Key(r, g, b)
	c.Lock()
	defer c.Unlock()
	if point, ok := c.Colors[key]; ok {
		return point
	}
	point := make([]float64, 3)
	h, s, l := hsl(uint8(float64(r)/256.0), uint8(float64(g)/256.0), uint8(float64(b)/256.0))
	point[0] = h // float64(r) / float64(0xffff)
	point[1] = s // float64(g) / float64(0xffff)
	point[2] = l // float64(b) / float64(0xffff)
	c.Colors[key] = point
	return point
}

func NewColorCache(size int) *ColorCache {
	return &ColorCache{
		Colors:     make(map[string][]float64, size),
		BufferPool: bpool.NewBufferPool(size),
	}
}
