package table

type ParsedByteTable []byte

func (pst ParsedByteTable) Get(i byte) byte {
	return byte(pst[i])
}

type ParsedWordTable []byte

func (pbt ParsedWordTable) Get(i byte) uint32 {
	data := pbt[4*uint(i) : 4*(uint(i)+1)]
	return (uint32(data[0]) << 24) | (uint32(data[1]) << 16) | (uint32(data[2]) << 8) | uint32(data[3])
}

func SerializeByteTable(t ByteTable) (out []byte) {
	for i := 0; i < 256; i++ {
		out = append(out, byte(t.Get(byte(i))))
	}

	return
}

func SerializeWordTable(t WordTable) (out []byte) {
	for i := 0; i < 256; i++ {
		val := t.Get(byte(i))
		out = append(out, byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
	}

	return
}
