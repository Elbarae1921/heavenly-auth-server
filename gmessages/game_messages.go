package gmessages

type LoginMessage struct {
	UserName string
	Password string
}

type RegisterMessage struct {
	UserName string
	Password string
	Email    string
}
