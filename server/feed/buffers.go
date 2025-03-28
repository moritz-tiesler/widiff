package feed

type RingBuffer[T any] interface {
	Read() T
	Write(T)
	Items() []T
}

type Buffer[T any] struct {
	items []T
	ip    int
	rp    int
}

func (b *Buffer[T]) Read() T {
	var v T
	if b.rp < 0 {
		return v
	}
	v = b.items[b.rp]
	b.rp++
	b.rp = b.rp % cap(b.items)
	return v
}

func (b *Buffer[T]) Write(e T) {
	b.items[b.ip] = e
	b.ip++
	b.ip = b.ip % cap(b.items)
}

func (b *Buffer[T]) Items() []T {
	return b.items
}

func NewBuffer[T any](cap int) *Buffer[T] {
	return &Buffer[T]{
		items: make([]T, cap, cap),
		ip:    0,
		rp:    0,
	}
}
