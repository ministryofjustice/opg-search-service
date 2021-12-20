package reindex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchIter(t *testing.T) {
	assert := assert.New(t)

	batch := &batchIter{start: 2, end: 25, size: 10}

	assert.True(batch.Next())
	assert.Equal(2, batch.From())
	assert.Equal(11, batch.To())

	assert.True(batch.Next())
	assert.Equal(12, batch.From())
	assert.Equal(21, batch.To())

	assert.True(batch.Next())
	assert.Equal(22, batch.From())
	assert.Equal(25, batch.To())

	assert.False(batch.Next())
}

func TestBatchIterBackwards(t *testing.T) {
	batch := &batchIter{start: 10, end: 0, size: 10}

	assert.False(t, batch.Next())
}

func TestBatchIterSingleItem(t *testing.T) {
	assert := assert.New(t)

	batch := &batchIter{start: 0, end: 0, size: 10}

	assert.True(batch.Next())
	assert.Equal(0, batch.From())
	assert.Equal(0, batch.To())

	assert.False(batch.Next())
}

func TestBatchIterOneBatch(t *testing.T) {
	assert := assert.New(t)

	batch := &batchIter{start: 5, end: 10, size: 10}

	assert.True(batch.Next())
	assert.Equal(5, batch.From())
	assert.Equal(10, batch.To())

	assert.False(batch.Next())
}
