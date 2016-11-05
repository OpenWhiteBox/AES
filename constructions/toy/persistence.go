package toy

import (
	"errors"

	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/matrix"
)

const fullSize = 11 * (128 + 1) * 16

// Serialize serializes a white-box construction into a byte slice.
func (constr *Construction) Serialize() []byte {
	out := make([]byte, 0)

	for _, round := range constr {
		for _, row := range round.Forwards {
			out = append(out, row...)
		}
		out = append(out, round.BlockAdditive[:]...)
	}

	return out
}

// Parse parses a byte array into a white-box construction. It returns an error if the byte slice isn't long enough.
func Parse(in []byte) (constr Construction, err error) {
	if len(in) != fullSize {
		err = errors.New("Parsing the key failed.")
		return
	}

	for round := 0; round < 11; round++ {
		forwards := matrix.Matrix{}
		constant := [16]byte{}

		for row := 0; row < 128; row++ {
			forwards = append(forwards, matrix.Row(in[:16]))
			in = in[16:]
		}
		copy(constant[:], in[:16])
		in = in[16:]

		constr[round] = encoding.NewBlockAffine(forwards, constant)
	}

	return
}
