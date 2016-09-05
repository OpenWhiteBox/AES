package chow

import (
	"testing"

	"bytes"
	"crypto/rand"

	"github.com/OpenWhiteBox/AES/constructions/chow"
	"github.com/OpenWhiteBox/AES/constructions/common"
)

func TestRecoverKey(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)

	constr, _, _ := chow.GenerateEncryptionKeys(
		key, key, common.IndependentMasks{common.RandomMask, common.RandomMask},
	)

	cand := RecoverKey(&constr)

	if !bytes.Equal(cand, key) {
		t.Fatalf("Recovered wrong key!\nreal=%x\ncand=%x", key, cand)
	}
}

// func TestMakeConstants(t *testing.T) {
//   MC := gfmatrix.Matrix{
//     gfmatrix.Row{2, 3, 1, 1},
//     gfmatrix.Row{1, 2, 3, 1},
//     gfmatrix.Row{1, 1, 2, 3},
//     gfmatrix.Row{3, 1, 1, 2},
//   }
//
//   MixColumns := [4][4]matrix.Matrix{}
//   UnMixColumns := [4][4]matrix.Matrix{}
//
//   for row := 0; row < 4; row++ {
//     for col := 0; col < 4; col++ {
//       lin := encoding.DecomposeByteLinear(encoding.NewByteMultiplication(MC[row][col]))
//
//       MixColumns[row][col] = lin.Forwards
//       UnMixColumns[row][col] = lin.Backwards
//     }
//   }
//
//   fmt.Println("mixColumns = [4][4]matrix.Matrix{")
//   for row := 0; row < 4; row++ {
//     fmt.Println("\t[4]matrix.Matrix{")
//     for col := 0; col < 4; col++ {
//       fmt.Println("\t\tmatrix.Matrix{")
//       for _, r := range MixColumns[row][col] {
//         fmt.Print("\t\t\tmatrix.Row{")
//         for _, c := range r[0:len(r)-1] {
//           fmt.Printf("0x%2.2x, ", c)
//         }
//         fmt.Printf("0x%2.2x},\n", r[len(r)-1])
//       }
//
//       fmt.Println("\t\t}")
//     }
//     fmt.Println("\t},")
//   }
//   fmt.Println("}\n")
//
//   fmt.Println("unMixColumns = [4][4]matrix.Matrix{")
//   for row := 0; row < 4; row++ {
//     fmt.Println("\t[4]matrix.Matrix{")
//     for col := 0; col < 4; col++ {
//       fmt.Println("\t\tmatrix.Matrix{")
//       for _, r := range UnMixColumns[row][col] {
//         fmt.Print("\t\t\tmatrix.Row{")
//         for _, c := range r[0:len(r)-1] {
//           fmt.Printf("0x%2.2x, ", c)
//         }
//         fmt.Printf("0x%2.2x},\n", r[len(r)-1])
//       }
//
//       fmt.Println("\t\t}")
//     }
//     fmt.Println("\t},")
//   }
//   fmt.Println("}")
// }
