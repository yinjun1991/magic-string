package magicstring

type bitSet struct {
	bits []byte
}

func (b *bitSet) add(n int) {
	b.bits[n>>5] |= 1 << (n & 31)
}

func (b *bitSet) has(n int) bool {
	bt := b.bits[n>>5] & (1 << (n & 31))
	return bt > 0
}

func newBitSet(bs *bitSet) *bitSet {
	bs2 := &bitSet{}

	if bs != nil {
		bs2.bits = make([]byte, len(bs.bits))
		copy(bs2.bits, bs.bits)
	}

	return bs2
}
