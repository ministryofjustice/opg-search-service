package reindex

import (
	"context"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockBulkClient struct {
	ops    []*elasticsearch.BulkOp
	result elasticsearch.BulkResult
	err    error
}

func (c *mockBulkClient) DoBulk(op *elasticsearch.BulkOp) (elasticsearch.BulkResult, error) {
	c.ops = append(c.ops, op)

	return c.result, c.err
}

func TestReindex(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	client := &mockBulkClient{
		result: elasticsearch.BulkResult{Successful: 1, Failed: 0},
	}
	r := &Reindexer{es: client}

	p := person.Person{ID: i64(1), Firstname: "A"}

	persons := make(chan person.Person, 1)
	persons <- p
	close(persons)

	expectedOp := elasticsearch.NewBulkOp("person")
	expectedOp.Index(p.Id(), p)

	result, err := r.reindex(ctx, persons)
	if assert.Nil(err) {
		assert.Equal(1, result.Successful)
		assert.Equal(0, result.Failed)
		assert.Empty(result.Errors)

		assert.Equal([]*elasticsearch.BulkOp{expectedOp}, client.ops)
	}
}
