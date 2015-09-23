package encoding

import (
	"../table"
)

type NibbleTable struct {
	In     Byte
	Out    Nibble
	Hidden table.Nibble
}

func (nt NibbleTable) Get(i byte) byte {
	return nt.Out.Encode(nt.Hidden.Get(nt.In.Decode(i)))
}

type ByteTable struct {
	In     Byte
	Out    Byte
	Hidden table.Byte
}

func (bt ByteTable) Get(i byte) byte {
	return bt.Out.Encode(bt.Hidden.Get(bt.In.Decode(i)))
}

type WordTable struct {
	In     Byte
	Out    Word
	Hidden table.Word
}

func (wt WordTable) Get(i byte) uint32 {
	return wt.Out.Encode(wt.Hidden.Get(wt.In.Decode(i)))
}
