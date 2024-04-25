package socketio_helper

import (
	"fmt"

	socketio "github.com/googollee/go-socket.io"
)

func InitializeSocketIO() {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("connected:", s.ID())
		return nil
	})

	server.OnEvent("/", "join-room", func(s socketio.Conn, msg string) {

	})

	server.OnEvent("/", "chat", func(s socketio.Conn, msg string) {

	})
}
