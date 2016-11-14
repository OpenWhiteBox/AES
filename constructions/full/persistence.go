package full

import (
	"errors"
)

// Serialize serializes a white-box construction into a byte slice.
func (constr *Construction) Serialize() []byte {
	out := make([]byte, 0)

	for _, round := range constr {
		round.serialize(&out)
	}

	return out
}

// Parse parses a byte array into a white-box construction. It returns an error if the byte slice isn't long enough.
func Parse(in []byte) (constr Construction, err error) {
	if len(in) != 1091178 {
		return constr, errors.New("key is the wrong size")
	}

	for i := 0; i < len(constr); i++ {
		constr[i], in = parseBlockAffine(in)
	}

	return
}
