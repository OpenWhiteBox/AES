// A table maps one byte to another primitive type (nibble, a second byte, ...).  They're not necessarily injective or
// surjective.  A series of byte tables can be composed indefinitely, but a nibble, word, or block table will terminate
// the series.
package table

type Nibble interface {
	Get(i byte) byte // Get(i byte) nibble
}

type Byte interface {
	Get(i byte) byte
}

type Word interface {
	Get(i byte) uint32
}

type Block interface {
	Get(i byte) [16]byte
}

type ComposedBytes []Byte

func (cb ComposedBytes) Get(i byte) byte {
	for j, _ := range cb {
		i = cb[j].Get(i)
	}

	return i
}

// ComposedToNibble isn't needed because you can use ComposedSmalls.

type ComposedToWord struct {
	Heads Byte
	Tails Word
}

func (cw ComposedToWord) Get(i byte) uint32 {
	return cw.Tails.Get(cw.Heads.Get(i))
}

type ComposedToBlock struct {
	Heads Byte
	Tails Block
}

func (cb ComposedToBlock) Get(i byte) [16]byte {
	return cb.Tails.Get(cb.Heads.Get(i))
}
