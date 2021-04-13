package person

type SearchRequest struct {
	Term        string   `json:"term"`
	Size        int      `json:"size,omitempty"`
	From        int      `json:"from"`
	PersonTypes []string `json:"person_types"`
}
