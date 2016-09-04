// Package bes implements the Big Encryption System, a variant of AES with more algebraic structure.
//
// From the abstract of the source paper:
//   One difficulty in the cryptanalysis of the Advanced Encryption Standard (AES) is the tension between operations in
//   the two fields GF(2^8) and GF(2). ... We define a new block cipher, the BES, that uses only simple operations in
//   GF(2^8). Yet, the AES can be regarded as being identical to the BES with a restricted message space and key space,
//   thus enabling the AES to be realised solely using simple algebraic operations in one field GF(2^8). This permits
//   the exploration of the AES within a broad and rich setting.
//
// "Essential Algebraic Structure Within the AES" by S. Murphy and M.J.B. Robshaw,
// http://www.isg.rhul.ac.uk/~sean/crypto.pdf
package bes

import (
	"github.com/OpenWhiteBox/primitives/gfmatrix"
	"github.com/OpenWhiteBox/primitives/number"
)

// Powers of x mod M(x).
var powx = [16]byte{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x1b, 0x36, 0x6c, 0xd8, 0xab, 0x4d, 0x9a, 0x2f}

type Construction struct {
	// A 128-byte BES key.
	Key gfmatrix.Row
}

// BlockSize returns the block size of BES. (Necessary to implement cipher.Block.)
func (constr Construction) BlockSize() int { return 128 }

// Encrypt encrypts the first block in src into dst. Dst and src may point at the same memory.
func (constr Construction) Encrypt(dst, src []byte) {
	state := gfmatrix.NewRow(128)
	for pos := 0; pos < 128; pos++ {
		state[pos] = number.ByteFieldElem(src[pos])
	}

	state = constr.encrypt(state)

	for pos := 0; pos < 128; pos++ {
		dst[pos] = byte(state[pos])
	}
}

// Decrypt decrypts the first block in src into dst. Dst and src may point at the same memory.
func (constr Construction) Decrypt(dst, src []byte) {
	state := gfmatrix.NewRow(128)
	for pos := 0; pos < 128; pos++ {
		state[pos] = number.ByteFieldElem(src[pos])
	}

	state = constr.decrypt(state)

	for pos := 0; pos < 128; pos++ {
		dst[pos] = byte(state[pos])
	}
}

func (constr *Construction) encrypt(in gfmatrix.Row) gfmatrix.Row {
	roundKeys := constr.StretchedKey()

	state := in.Add(roundKeys[0])

	for i := 1; i <= 9; i++ {
		state = constr.subBytes(state)
		state = round.Mul(state)
		state = state.Add(roundConst).Add(roundKeys[i])
	}

	state = constr.subBytes(state)
	state = lastRound.Mul(state)
	state = state.Add(roundConst).Add(roundKeys[10])

	return state
}

func (constr *Construction) decrypt(in gfmatrix.Row) gfmatrix.Row {
	roundKeys := constr.StretchedKey()

	state := in.Add(roundConst).Add(roundKeys[10])
	state = firstRound.Mul(state)
	state = constr.subBytes(state)

	for i := 9; i >= 1; i-- {
		state = state.Add(roundConst).Add(roundKeys[i])
		state = unRound.Mul(state)
		state = constr.subBytes(state)
	}

	state = state.Add(roundKeys[0])

	return state
}

// StretchedKey implements BES' key schedule. It returns the 11 round keys derived from the master key.
func (constr *Construction) StretchedKey() [11]gfmatrix.Row {
	var (
		i         int = 0
		stretched [4 * 11]gfmatrix.Row
		split     [11]gfmatrix.Row
	)

	for ; i < 4; i++ {
		stretched[i] = constr.Key[32*i : 32*(i+1)]
	}

	for ; i < (4 * 11); i++ {
		temp := stretched[i-1]

		if (i % 4) == 0 {
			temp = append(temp[8:], temp[:8]...)
			temp = constr.subBytes(temp)
			temp = wordSubBytes.Mul(temp).Add(wordSubBytesConst)
			temp = temp.Add(Expand([]byte{powx[i/4-1], 0, 0, 0}))
		}

		stretched[i] = stretched[i-4].Add(temp)
	}

	for j := 0; j < 11; j++ {
		split[j] = gfmatrix.NewRow(128)

		for k := 0; k < 4; k++ {
			copy(split[j][32*k:32*(k+1)], stretched[4*j+k])
		}
	}

	return split
}

// subBytes rewrites each byte of the state with the S-Box. unSubBytes and subBytes are the same.
func (constr *Construction) subBytes(in gfmatrix.Row) gfmatrix.Row {
	out := gfmatrix.NewRow(in.Size())

	for pos := 0; pos < in.Size(); pos++ {
		out[pos] = in[pos].Invert()
	}

	return out
}
