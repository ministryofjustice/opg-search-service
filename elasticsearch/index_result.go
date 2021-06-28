package elasticsearch

type IndexResult struct {
	Id         int64  `json:"id"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type BulkResult struct {
	StatusCode int              `json:"statusCode"`
	Message    string           `json:"message"`
	Results    []BulkResultItem `json:"results"`
}

type BulkResultItem struct {
	ID         string `json:"id"`
	StatusCode int    `json:"statusCode"`
}
