package table

type NibbleTable interface {
	Get(i byte) byte
}

type ByteTable interface {
	Get(i byte) byte
}

type WordTable interface {
	Get(i byte) uint32
}

type ComposedToWordTable struct {
	Heads ByteTable
	Tails WordTable
}

func (cwt ComposedToWordTable) Get(i byte) uint32 {
	return cwt.Tails.Get(cwt.Heads.Get(i))
}

type ComposedSmallTables []ByteTable

func (cst ComposedSmallTables) Get(i byte) byte {
	for j, _ := range cst {
		i = cst[j].Get(i)
	}

	return i
}
