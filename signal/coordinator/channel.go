package coordinator

// Channel manages users.
type Channel struct {
	//TODO(window9u): we should add lock for channels
	users map[string]*User
}
