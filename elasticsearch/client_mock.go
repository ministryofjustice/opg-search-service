package elasticsearch

import "github.com/stretchr/testify/mock"

type MockESClient struct {
	mock.Mock
}

func (m *MockESClient) Index(i Indexable) *IndexResult {
	args := m.Called(i)
	return args.Get(0).(*IndexResult)
}

func (m *MockESClient) Search(requestBody map[string]interface{}, dataType Indexable) (*[]string, error) {
	args := m.Called(requestBody, dataType)
	return args.Get(0).(*[]string), args.Error(1)
}

func (m *MockESClient) CreateIndex(i Indexable) (bool, error) {
	args := m.Called(i)
	return args.Get(0).(bool), args.Error(1)
}