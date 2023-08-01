package index

import "github.com/ministryofjustice/opg-search-service/internal/response"

type mockIndexable struct {
	id string
}

func (m mockIndexable) Id() string {
	return m.id
}

type mockValidatable struct {
	error []response.Error
	items []Indexable
}

func (m mockValidatable) Validate() []response.Error {
	return m.error
}

func (m mockValidatable) Items() []Indexable {
	return m.items
}
