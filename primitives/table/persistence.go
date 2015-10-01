// For efficiently persisting tables in storage.
package table

type ParsedByte []byte

func (pst ParsedByte) Get(i byte) byte {
	return byte(pst[i])
}

type ParsedWord []byte

func (pbt ParsedWord) Get(i byte) (out [4]byte) {
	data := pbt[4*uint(i) : 4*(uint(i)+1)]
	copy(out[:], data)

	return
}

type ParsedBlock []byte

func (pbt ParsedBlock) Get(i byte) (out [16]byte) {
	copy(out[:], pbt[16*uint(i):16*(uint(i)+1)])
	return
}

func SerializeByte(t Byte) (out []byte) {
	for i := 0; i < 256; i++ {
		out = append(out, byte(t.Get(byte(i))))
	}

	return
}

func SerializeWord(t Word) (out []byte) {
	for i := 0; i < 256; i++ {
		val := t.Get(byte(i))
		out = append(out, val[:]...)
	}

	return
}

func SerializedBlock(t Block) (out []byte) {
	for i := 0; i < 256; i++ {
		res := t.Get(byte(i))
		out = append(out, res[:]...)
	}

	return
}
