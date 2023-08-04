package index

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLogger struct{}

func (*mockLogger) Printf(s string, args ...interface{}) {}

type mockDB struct {
	mock.Mock
}

func (m *mockDB) QueryIDRange(ctx context.Context) (min, max int, err error) {
	args := m.Called(ctx)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *mockDB) QueryByID(ctx context.Context, results chan<- Indexable, from, to int) error {
	args := m.Called(ctx, results, from, to)
	return args.Error(0)
}

func (m *mockDB) QueryFromDate(ctx context.Context, results chan<- Indexable, from time.Time) error {
	args := m.Called(ctx, results, from)
	return args.Error(0)
}

type mockClient struct {
	mock.Mock
}

func (m *mockClient) DoBulk(ctx context.Context, op *elasticsearch.BulkOp) (elasticsearch.BulkResult, error) {
	args := m.Called(ctx, op)
	return args.Get(0).(elasticsearch.BulkResult), args.Error(1)
}

func TestAll(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	first := mockIndexable{id: 1}
	second := mockIndexable{id: 2}

	db := &mockDB{}
	db.
		On("QueryIDRange", ctx).
		Return(1, 3, nil)
	db.
		On("QueryByID", ctx, mock.Anything, 1, 3).
		Run(func(args mock.Arguments) {
			ch := args.Get(1).(chan<- Indexable)
			ch <- first
			ch <- second
		}).
		Return(nil)

	bulkOp := elasticsearch.NewBulkOp("whatever")
	bulkOp.Index(1, first)
	bulkOp.Index(2, second)

	client := &mockClient{}
	client.
		On("DoBulk", ctx, bulkOp).
		Return(elasticsearch.BulkResult{Successful: 2, Failed: 0}, nil)

	indexer := New(client, &mockLogger{}, db, "whatever")

	result, err := indexer.All(ctx, 10)
	assert.Nil(err)
	assert.Equal(&Result{Successful: 2}, result)

	mock.AssertExpectationsForObjects(t, db, client)
}

func TestByID(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	first := mockIndexable{id: 1}
	second := mockIndexable{id: 2}

	db := &mockDB{}
	db.
		On("QueryByID", ctx, mock.Anything, 3, 5).
		Run(func(args mock.Arguments) {
			ch := args.Get(1).(chan<- Indexable)
			ch <- first
			ch <- second
		}).
		Return(nil)

	bulkOp := elasticsearch.NewBulkOp("whatever")
	bulkOp.Index(1, first)
	bulkOp.Index(2, second)

	client := &mockClient{}
	client.
		On("DoBulk", ctx, bulkOp).
		Return(elasticsearch.BulkResult{Successful: 2, Failed: 0}, nil)

	indexer := New(client, &mockLogger{}, db, "whatever")

	result, err := indexer.ByID(ctx, 3, 5, 10)
	assert.Nil(err)
	assert.Equal(&Result{Successful: 2}, result)

	mock.AssertExpectationsForObjects(t, db, client)
}

func TestFromDate(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	fromDate := time.Now()

	first := mockIndexable{id: 1}
	second := mockIndexable{id: 2}

	db := &mockDB{}
	db.
		On("QueryFromDate", ctx, mock.Anything, fromDate).
		Run(func(args mock.Arguments) {
			ch := args.Get(1).(chan<- Indexable)
			ch <- first
			ch <- second
		}).
		Return(nil)

	bulkOp := elasticsearch.NewBulkOp("whatever")
	bulkOp.Index(1, first)
	bulkOp.Index(2, second)

	client := &mockClient{}
	client.
		On("DoBulk", ctx, bulkOp).
		Return(elasticsearch.BulkResult{Successful: 2, Failed: 0}, nil)

	indexer := New(client, &mockLogger{}, db, "whatever")

	result, err := indexer.FromDate(ctx, fromDate, 10)
	assert.Nil(err)
	assert.Equal(&Result{Successful: 2}, result)

	mock.AssertExpectationsForObjects(t, db, client)
}
