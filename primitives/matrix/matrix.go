package matrix

type ByteMatrix [8]byte // One byte is one row

func dotProduct(e, f byte) bool {
  weight := 0
  x := e & f

  for i := uint(0); i < 8; i++ {
    if x&(1<<i) > 0 {
      weight++
    }
  }

  if weight%2 == 0 {
    return false
  } else {
    return true
  }
}

func (e ByteMatrix) Mul(f byte) (out byte) {
  for i := uint(0); i < 8; i++ {
    if dotProduct(e[i], f) {
      out += 1 << i
    }
  }

  return
}
