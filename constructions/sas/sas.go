// Package sas implements a generic SAS block cipher.
package sas

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

type Construction struct {
	First    encoding.Block
	Linear   matrix.Matrix
	Constant [16]byte
	Last     encoding.Block
}

func (constr Construction) Encrypt(dst, src []byte) {
	temp := [16]byte{}
	copy(temp[:], src)

	// First S-Box layer.
	temp = constr.First.Encode(temp)

	// Affine layer.
	copy(temp[:], constr.Linear.Mul(matrix.Row(temp[:])))
	for i := 0; i < 16; i++ {
		temp[i] = temp[i] ^ constr.Constant[i]
	}

	// Second S-Box layer.
	temp = constr.Last.Encode(temp)

	copy(dst, temp[:])
}

func (constr Construction) Decrypt(dst, src []byte) {
	temp := [16]byte{}
	copy(temp[:], src)

	// First S-Box layer.
	temp = constr.Last.Decode(temp)

	// Affine layer.
	for i := 0; i < 16; i++ {
		temp[i] = temp[i] ^ constr.Constant[i]
	}
	linearInv, _ := constr.Linear.Invert()
	copy(temp[:], linearInv.Mul(matrix.Row(temp[:])))

	// Second S-Box layer.
	temp = constr.First.Decode(temp)

	copy(dst, temp[:])
}
