package buffer

import (
	"sync"
	"unsafe"
)

var pool = sync.Pool{
	New: func() any {
		return &Bytes{B: make([]byte, 0, 1024)}
	},
}

func Get() *Bytes {
	return pool.Get().(*Bytes)
}

func Put(b *Bytes) {
	if b.Len() > 512*1024 { // Avoid large buffer.
		b = nil

		return
	}

	b.Reset()
	pool.Put(b)
}

// Bytes is a simple buffer.
// It is unsafe, SHOULD not modify existing bytes.
type Bytes struct {
	B []byte
}

func (bs *Bytes) Reset() {
	bs.B = bs.B[:0]
}

func (bs *Bytes) String() string {
	return string(bs.B)
}

func (bs *Bytes) Bytes() []byte {
	return bs.B
}

func (bs *Bytes) Write(p []byte) (n int, err error) {
	bs.B = append(bs.B, p...)

	return len(p), nil
}

// Unsafe!!!
func (bs *Bytes) WriteString(s string) (n int, err error) {
	b := unsafe.Slice(unsafe.StringData(s), len(s))

	return bs.Write(b)
}

// EnsureRemaining ensures the buffer has space for at least `atLeast`
// additional bytes beyond the current length (i.e., remaining capacity).
// It grows the buffer if necessary using an amortized growth strategy.
func (bs *Bytes) EnsureRemaining(atLeast int) {
	if atLeast <= 0 {
		return
	}

	// Calculate the minimum total capacity required.
	// needCap = current_length + required_remaining_capacity
	needCap := len(bs.B) + atLeast
	if cap(bs.B) >= needCap {
		// Current capacity is already sufficient.
		return
	}

	// --- Need to grow ---

	// Determine the new capacity.
	// Strategy: Double the existing capacity, but make sure it's at least needCap.
	// This amortizes the cost of allocations over time.
	newCap := max(cap(bs.B)*2, needCap)

	// Allocate a new slice with the current length and the calculated new capacity.
	// Note: We create it with the *current length*, not zero length.
	newB := make([]byte, len(bs.B), newCap)

	// Copy the existing data from the old buffer to the new buffer.
	copy(newB, bs.B) // copy is efficient

	// Replace the buffer's internal slice with the new one.
	bs.B = newB
}

func (bs *Bytes) Remaining() int {
	return cap(bs.B) - len(bs.B)
}

func (bs *Bytes) Len() int {
	return len(bs.B)
}

func (bs *Bytes) Cap() int {
	return cap(bs.B)
}
