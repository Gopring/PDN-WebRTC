package message

type Push struct {
	ChannelID string
	UserID    string
	SDP       string
}

type Pull struct {
	ChannelID string
	UserID    string
	SDP       string
}
