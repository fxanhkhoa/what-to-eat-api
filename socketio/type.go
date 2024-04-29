package socketio_helper

type SocketioJoinRoom struct {
	RoomID string `json:"roomID"`
}

type SocketioDishVoteUpdate struct {
	Slug     string `json:"slug"`
	MyName   string `json:"myName"`
	UserID   string `json:"userID"`
	IsVoting bool   `json:"isVoting"`
}
