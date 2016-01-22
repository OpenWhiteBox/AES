package equivalence

import (
	// "fmt"
	"testing"

	// "github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
)

type InvertEncoding struct{}

func (ie InvertEncoding) Encode(in byte) byte { return byte(number.ByteFieldElem(in).Invert()) }
func (ie InvertEncoding) Decode(in byte) byte { return ie.Encode(in) }

// func TestLinearEquivalence(t *testing.T) {
// 	f := InvertEncoding{}
// 	eqs := LinearEquivalence(f, f, 2041)
// 	fmt.Println(len(eqs))
// 	// for _, eq := range eqs {
// 	// 	fA := encoding.ComposedBytes{encoding.ByteLinear{eq.A, nil}, f}
// 	// 	Bf := encoding.ComposedBytes{f, encoding.ByteLinear{eq.B, nil}}
// 	//
// 	// 	for x := 0; x < 256; x++ {
// 	// 		if fA.Encode(byte(x)) != Bf.Encode(byte(x)) {
// 	// 			t.Fatalf("Bad equivalence!")
// 	// 		}
// 	// 	}
// 	// }
// }

func BenchmarkLearn(b *testing.B) {
	f, g := InvertEncoding{}, InvertEncoding{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		A, B := NewMatrix(), NewMatrix()
		A.Assert(matrix.Row{0x01}, matrix.Row{0x06})
		A.Assert(matrix.Row{0x02}, matrix.Row{0xdf})

		learn(f, g, A, B)
	}
}
