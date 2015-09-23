package table

type Nibble interface {
	Get(i byte) byte
}

type Byte interface {
	Get(i byte) byte
}

type Word interface {
	Get(i byte) uint32
}

type ComposedToWord struct {
	Heads Byte
	Tails Word
}

func (cw ComposedToWord) Get(i byte) uint32 {
	return cw.Tails.Get(cw.Heads.Get(i))
}

type ComposedSmalls []Byte

func (cs ComposedSmalls) Get(i byte) byte {
	for j, _ := range cs {
		i = cs[j].Get(i)
	}

	return i
}
