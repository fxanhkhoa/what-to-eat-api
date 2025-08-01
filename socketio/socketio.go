package socketio

import (
	"encoding/json"
	"fmt"
	"what-to-eat/be/model"
	socketio_service "what-to-eat/be/socketio/service"

	"github.com/labstack/echo/v4"
	"github.com/zishang520/socket.io/socket"
)

func InitializeSocketIO(e *echo.Echo) *socket.Server {
	// Configure Socket.IO server for WebSocket transport
	io := socket.NewServer(nil, nil)

	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		client.On("join-room", func(datas ...any) {
			jsonStr, err := json.Marshal(datas[0])
			if err != nil {
				fmt.Println(err)
			}
			var socketioJoinRoomData model.SocketioJoinRoom
			if err := json.Unmarshal(jsonStr, &socketioJoinRoomData); err != nil {
				fmt.Println(err)
			}

			client.Join(socket.Room(socketioJoinRoomData.RoomID))
		})
		client.On("dish-vote-update", func(data ...any) {
			updated, socketioJoinRoomData := socketio_service.ProcessDishVoteUpdate(data...)

			io.To(socket.Room(socketioJoinRoomData.RoomID)).Emit("dish-vote-update-client", updated)
		})
		client.On("disconnect", func(...any) {
		})
	})

	// Serve Socket.IO over WebSocket and HTTP
	e.Any("/socket.io/*", func(c echo.Context) error {
		// The Socket.IO server handles both WebSocket upgrades and HTTP polling
		handler := io.ServeHandler(nil)
		handler.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	})

	return io
}
