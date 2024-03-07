package hb

// Person is a single user.
type Person struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}
