package elasticsearch

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockESClient struct {
	mock.Mock
}

func (m *MockESClient) DoBulk(ctx context.Context, op *BulkOp) (BulkResult, error) {
	args := m.Called(ctx, op)
	return args.Get(0).(BulkResult), args.Error(1)
}

func (m *MockESClient) Search(ctx context.Context, indexName []string, requestBody map[string]interface{}) (*SearchResult, error) {
	args := m.Called(ctx, indexName, requestBody)
	return args.Get(0).(*SearchResult), args.Error(1)
}

func (m *MockESClient) Delete(ctx context.Context, indexName []string, requestBody map[string]interface{}) (*DeleteResult, error) {
	args := m.Called(ctx, indexName, requestBody)
	return args.Get(0).(*DeleteResult), args.Error(1)
}

func (m *MockESClient) CreateIndex(ctx context.Context, name string, config []byte, force bool) error {
	args := m.Called(ctx, name, config, force)
	return args.Error(0)
}

func (m *MockESClient) ResolveAlias(ctx context.Context, alias string) (string, error) {
	args := m.Called(ctx, alias)
	return args.String(0), args.Error(1)
}

func (m *MockESClient) CreateAlias(ctx context.Context, alias string, index string) error {
	args := m.Called(ctx, alias, index)
	return args.Error(0)
}
