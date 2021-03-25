package response

type IndexResponse struct {
	Results []struct {
		Id         string
		StatusCode int
	}
}
