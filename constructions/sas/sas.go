// Package sas implements a generic SAS block cipher.
//
// An affine layer, denoted by an A, treats its input as an element of GF(2)^n and applies a fixed, invertible affine
// transformation over this space. An S-box layer, denoted by an S, applies possibly independent 8-bit S-boxes to
// consecutive chunks of its input. Therefore, an SAS block cipher is a product cipher with an affine layer between two
// S-box layers.
//
// An efficient cryptanalysis of SAS block ciphers is implemented in the cryptanalysis/sas package.
//
// "Structural Cryptanalysis of SASAS" by Alex Biryukov and Adi Shamir,
// https://www.iacr.org/archive/eurocrypt2001/20450392.pdf
package sas

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
)

type Construction struct {
	First  encoding.ConcatenatedBlock
	Affine encoding.BlockAffine
	Last   encoding.ConcatenatedBlock
}

func (constr Construction) cipher() encoding.Block {
	return encoding.ComposedBlocks{
		constr.First,
		constr.Affine,
		constr.Last,
	}
}

// BlockSize returns the block size of AES. (Necessary to implement cipher.Block.)
func (constr Construction) BlockSize() int { return 16 }

// Encrypt encrypts the first block in src into dst. Dst and src may point at the same memory.
func (constr Construction) Encrypt(dst, src []byte) {
	temp := [16]byte{}
	copy(temp[:], src)

	temp = constr.cipher().Encode(temp)

	copy(dst, temp[:])
}

// Decrypt decrypts the first block in src into dst. Dst and src may point at the same memory.
func (constr Construction) Decrypt(dst, src []byte) {
	temp := [16]byte{}
	copy(temp[:], src)

	temp = constr.cipher().Decode(temp)

	copy(dst, temp[:])
}
