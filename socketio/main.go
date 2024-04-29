package socketio_helper

import (
	"encoding/json"
	"fmt"

	"github.com/zishang520/socket.io/socket"
)

func InitializeSocketIO() *socket.Server {
	io := socket.NewServer(nil, nil)

	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("join-room", func(datas ...any) {
			jsonStr, err := json.Marshal(datas[0])
			if err != nil {
				fmt.Println(err)
			}
			var socketioJoinRoomData SocketioJoinRoom
			if err := json.Unmarshal(jsonStr, &socketioJoinRoomData); err != nil {
				fmt.Println(err)
			}

			client.Join(socket.Room(socketioJoinRoomData.RoomID))
		})
		client.On("dish-vote-update", func(data ...any) {
			updated, socketioJoinRoomData := ProcessDishVoteUpdate(data...)

			io.To(socket.Room(socketioJoinRoomData.RoomID)).Emit("dish-vote-update-client", updated)
		})
		client.On("disconnect", func(...any) {
		})
	})

	return io
}
