package index

import "github.com/ministryofjustice/opg-search-service/internal/response"

type mockIndexable struct {
	id int
}

func (m mockIndexable) Id() int64 {
	return int64(m.id)
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
