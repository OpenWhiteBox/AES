package chow

import (
	"testing"

	"github.com/OpenWhiteBox/AES/constructions/chow"
	"github.com/OpenWhiteBox/AES/primitives/table"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

func testConstruction() (chow.Construction, [16]byte) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	constr, _, _ := chow.GenerateKeys(key, seed)

	return constr, key
}

// func TestRecoverKey(t *testing.T) {
// 	key := test_vectors.AESVectors[50].Key
// 	seed := test_vectors.AESVectors[51].Key
//
//   fmt.Println("a")
// 	constr, _, _ := chow.GenerateKeys(key, seed)
//   fmt.Println("b")
// 	fmt.Println(RecoverKey(constr))
// }

func TestFindBasisAndSort(t *testing.T) {
	constr, _ := testConstruction()
	S := GenerateS(constr)

	basis := FindBasisAndSort(S)

	for i, perm := range S {
		vect := [256]byte{}
		copy(vect[:], table.SerializeByte(FunctionFromBasis(i, basis)))

		if vect != perm {
			t.Fatalf("FindBasisAndSort, position #%v is wrong!\n%v\n%v", i, vect, perm)
		}
	}
}
