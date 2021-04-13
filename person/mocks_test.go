package person

import (
	"github.com/stretchr/testify/mock"
	"opg-search-service/elasticsearch"
)

type MockESClient struct {
	mock.Mock
}

func (m *MockESClient) Index(i elasticsearch.Indexable) *elasticsearch.IndexResult {
	args := m.Called(i)
	return args.Get(0).(*elasticsearch.IndexResult)
}

func (m *MockESClient) Search(requestBody map[string]interface{}, dataType elasticsearch.Indexable) ([]string, error) {
	args := m.Called(requestBody, dataType)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockESClient) CreateIndex(i elasticsearch.Indexable) (bool, error) {
	args := m.Called(i)
	return args.Get(0).(bool), args.Error(1)
}
