package elasticsearch

import "github.com/stretchr/testify/mock"

type MockESClient struct {
	mock.Mock
}

func (m *MockESClient) DoBulk(op *BulkOp) (BulkResult, error) {
	args := m.Called(op)
	return args.Get(0).(BulkResult), args.Error(1)
}

func (m *MockESClient) Search(indexName string, requestBody map[string]interface{}) (*SearchResult, error) {
	args := m.Called(indexName, requestBody)
	return args.Get(0).(*SearchResult), args.Error(1)
}

func (m *MockESClient) CreateIndex(name string, config []byte, force bool) error {
	args := m.Called(name, config, force)
	return args.Error(0)
}

func (m *MockESClient) ResolveAlias(alias string) (string, error) {
	args := m.Called(alias)
	return args.String(0), args.Error(1)
}

func (m *MockESClient) CreateAlias(alias string, index string) error {
	args := m.Called(alias, index)
	return args.Error(0)
}
