package elasticsearch

type IndexResult struct {
	Id         int64  `json:"id"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}
