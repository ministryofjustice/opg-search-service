package elasticsearch

import "github.com/stretchr/testify/mock"

type MockESClient struct {
	mock.Mock
}

func (m *MockESClient) DoBulk(op *BulkOp) (BulkResult, error) {
	args := m.Called(op)
	return args.Get(0).(BulkResult), args.Error(1)
}

func (m *MockESClient) Search(requestBody map[string]interface{}, dataType Indexable) (*SearchResult, error) {
	args := m.Called(requestBody, dataType)
	return args.Get(0).(*SearchResult), args.Error(1)
}

func (m *MockESClient) CreateIndex(i Indexable) (bool, error) {
	args := m.Called(i)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockESClient) IndexExists(i Indexable) (bool, error) {
	args := m.Called(i)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockESClient) DeleteIndex(i Indexable) error {
	args := m.Called(i)
	return args.Error(0)
}
