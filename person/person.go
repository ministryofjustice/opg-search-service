package person

type Person struct {
	Id        int
	FirstName string
	LastName  string
}

func (p Person) GetIndexName() string {
	return "person"
}

func (p Person) GetJson() string {
	return `{"FirstName": "` + p.FirstName + `", "LastName": "` + p.LastName + `"}`
}
