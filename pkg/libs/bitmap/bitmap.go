package bitmap

import (
	"fmt"
)

type Bitmap struct {
	bits     []uint64
	position int // current assigned bit position
	size     int // size of the bitmap
}

func NewBitmap(n int) *Bitmap {
	size := (n + 63) / 64
	return &Bitmap{
		bits:     make([]uint64, size),
		position: 0,
		size:     n,
	}
}

func (b *Bitmap) Set(pos int) error {
	if pos < 0 || pos >= b.size {
		return fmt.Errorf("position out of range")
	}
	index := pos / 64
	offset := pos % 64
	b.bits[index] |= 1 << offset
	return nil
}

// Clear, and set the first pos to be 0
func (b *Bitmap) Clear(pos int) error {
	if pos < 0 || pos >= b.size {
		return fmt.Errorf("position out of range")
	}
	index := pos / 64
	offset := pos % 64
	b.bits[index] &^= 1 << offset
	return nil
}

// Get obtains the value of bit，return true if it is 1，false if it is 0
func (b *Bitmap) Get(pos int) (bool, error) {
	if pos < 0 || pos >= b.size {
		return false, fmt.Errorf("position out of range")
	}
	index := pos / 64
	offset := pos % 64
	return (b.bits[index] & (1 << offset)) != 0, nil
}

func (b *Bitmap) FindFirstUnset() int {
	n := b.size
	for i := 0; i < n; i++ {
		pos := (b.position + i) % n
		if value, _ := b.Get(pos); !value {
			b.position = (pos + 1) % n
			return pos
		}
	}
	return -1
}

func (b *Bitmap) FindFirstUnsetAndSet() int {
	i := b.FindFirstUnset()
	if i != -1 {
		b.Set(i)
	}
	return i
}
