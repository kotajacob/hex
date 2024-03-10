package hb

// Person is a single user.
type Person struct {
	ActorID     string `json:"actor_id"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Admin       bool   `json:"admin"`
	Local       bool   `json:"local"`
}
