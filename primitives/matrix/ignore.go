// Ignore functions for controlling matrix generation and manipulation.
package matrix

type RowIgnore func(int) bool
type ByteIgnore func(int, int) bool

func IgnoreNoRows(row int) bool {
	return false
}

func IgnoreNoBytes(row, col int) bool {
	return false
}

func IgnoreRows(positions ...int) RowIgnore {
	return func(row int) bool {
		pos := row / 8

		for _, cand := range positions {
			if pos == cand {
				return true
			}
		}

		return false
	}
}

func IgnoreBytes(positions ...int) ByteIgnore {
	return func(row, col int) bool {
		for _, pos := range positions {
			if row == pos || col == pos {
				return true
			}
		}

		return false
	}
}
