package saes

import (
	"testing"
)

var key = [16]byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33}

// func TestShiftRows(t *testing.T) {
// 	in := [16]byte{99, 202, 183, 4, 9, 83, 208, 81, 205, 96, 224, 231, 186, 112, 225, 140}
// 	out := [16]byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
//
// 	constr := Construction{key}
// 	cand := constr.shiftRows(in)
//
// 	for i := 0; i < 16; i++ {
// 		if out[i] != cand[i] {
// 			t.Fatalf("Byte %v is wrong! %v != %v", i, out[i], cand[i])
// 		}
// 	}
// }
//
// func TestMixColumns(t *testing.T) {
// 	in := [16]byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
// 	out := [16]byte{95, 114, 100, 21, 87, 245, 188, 146, 247, 190, 59, 41, 29, 185, 249, 26}
//
// 	constr := Construction{key}
// 	cand := constr.mixColumns(in)
//
// 	for i := 0; i < 16; i++ {
// 		if out[i] != cand[i] {
// 			t.Fatalf("Byte %v is wrong! %v != %v", i, out[i], cand[i])
// 		}
// 	}
// }
//
func TestEncrypt(t *testing.T) {
	testKey := [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	table := GenerateTables(testKey)

	in := [16]byte{0, 17, 34, 51, 68, 85, 102, 119, 136, 153, 170, 187, 204, 221, 238, 255}
	out := [16]byte{105, 196, 224, 216, 106, 123, 4, 48, 216, 205, 183, 128, 112, 180, 197, 90}

	constr := Construction{table}
	cand := constr.Encrypt(in)

	for i := 0; i < 16; i++ {
		if out[i] != cand[i] {
			t.Fatalf("Byte %v is wrong! %v != %v", i, out[i], cand[i])
		}
	}
}
