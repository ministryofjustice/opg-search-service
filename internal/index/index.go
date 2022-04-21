package index

import "github.com/ministryofjustice/opg-search-service/internal/response"

type Validatable interface {
	Validate() []response.Error
	Items() []Indexable
}

type Indexable interface {
	Id() int64
}
