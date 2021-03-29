package person

type Person struct {
	id        int
	FirstName string
	LastName  string
}

func (p Person) Id() int {
	return p.id
}

func (p Person) SetId(id int) {
	p.id = id
}

func (p Person) IndexName() string {
	return "person"
}

func (p Person) Json() string {
	return `{"FirstName": "` + p.FirstName + `", "LastName": "` + p.LastName + `"}`
}
