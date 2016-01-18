package random

import (
	"testing"
)

func TestStream(t *testing.T) {
	source := NewSource("Random Tests", make([]byte, 16))

	seed1 := make([]byte, 16)
	seed2 := make([]byte, 16)
	seed2[0] = 0x01

	stream1 := source.Stream(seed1)
	stream2 := source.Stream(seed1)
	stream3 := source.Stream(seed2)

	out1, out2, out3 := [16]byte{}, [16]byte{}, [16]byte{}

	stream1.Read(out1[:])
	stream2.Read(out2[:])
	stream3.Read(out3[:])

	if out1 != out2 {
		t.Fatalf("Two streams with same seed didn't give same output!")
	}

	if out1 == out3 {
		t.Fatalf("Two streams with different seed gave same output!")
	}
}
