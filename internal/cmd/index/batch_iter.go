package index

// batchIter iterates over the range start-end (inclusive) giving batches of
// size
type batchIter struct {
	start, end, size int
	i                int
	first            bool
}

func (i *batchIter) Next() bool {
	if !i.first {
		i.first = true
		i.i = i.start
	} else {
		i.i += i.size
	}

	return i.i <= i.end
}

func (i *batchIter) From() int {
	return i.i
}

func (i *batchIter) To() int {
	to := i.i + i.size - 1
	if to >= i.end {
		return i.end
	}

	return to
}
