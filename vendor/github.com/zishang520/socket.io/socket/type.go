package socket

import (
	"time"

	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/socket.io-go-parser/parser"
)

type Adapter interface {
	events.EventEmitter

	Rooms() *types.Map[Room, *types.Set[SocketId]]
	Sids() *types.Map[SocketId, *types.Set[Room]]
	Nsp() NamespaceInterface

	// To be overridden
	Init()

	// To be overridden
	Close()

	// Returns the number of Socket.IO servers in the cluster
	ServerCount() int64

	// Adds a socket to a list of room.
	AddAll(SocketId, *types.Set[Room])

	// Removes a socket from a room.
	Del(SocketId, Room)

	// Removes a socket from all rooms it's joined.
	DelAll(SocketId)

	SetBroadcast(func(*parser.Packet, *BroadcastOptions))
	GetBroadcast() func(*parser.Packet, *BroadcastOptions)
	// Broadcasts a packet.
	//
	// Options:
	//  - `Flags` {*BroadcastFlags} flags for this packet
	//  - `Except` {*types.Set[Room]} sids that should be excluded
	//  - `Rooms` {*types.Set[Room]} list of rooms to broadcast to
	Broadcast(*parser.Packet, *BroadcastOptions)

	// Broadcasts a packet and expects multiple acknowledgements.
	//
	// Options:
	//  - `Flags` {*BroadcastFlags} flags for this packet
	//  - `Except` {*types.Set[Room]} sids that should be excluded
	//  - `Rooms` {*types.Set[Room]} list of rooms to broadcast to
	BroadcastWithAck(*parser.Packet, *BroadcastOptions, func(uint64), func([]any, error))

	// Gets a list of sockets by sid.
	Sockets(*types.Set[Room]) *types.Set[SocketId]

	// Gets the list of rooms a given socket has joined.
	SocketRooms(SocketId) *types.Set[Room]

	// Returns the matching socket instances
	FetchSockets(*BroadcastOptions) []SocketDetails

	// Makes the matching socket instances join the specified rooms
	AddSockets(*BroadcastOptions, []Room)

	// Makes the matching socket instances leave the specified rooms
	DelSockets(*BroadcastOptions, []Room)

	// Makes the matching socket instances disconnect
	DisconnectSockets(*BroadcastOptions, bool)

	// Send a packet to the other Socket.IO servers in the cluster
	ServerSideEmit([]any) error

	// Save the client session in order to restore it upon reconnection.
	PersistSession(*SessionToPersist)

	// Restore the session and find the packets that were missed by the client.
	RestoreSession(PrivateSessionId, string) (*Session, error)
}

type AdapterConstructor interface {
	New(NamespaceInterface) Adapter
}

type SocketDetails interface {
	Id() SocketId
	Handshake() *Handshake
	Rooms() *types.Set[Room]
	Data() any
}

type NamespaceInterface interface {
	EventEmitter() *StrictEventEmitter

	On(string, ...events.Listener) error
	Once(string, ...events.Listener) error
	EmitReserved(string, ...any)
	EmitUntyped(string, ...any)
	Listeners(string) []events.Listener

	Sockets() *types.Map[SocketId, *Socket]
	Server() *Server
	Adapter() Adapter
	Name() string
	Ids() uint64

	// Sets up namespace middleware.
	Use(func(*Socket, func(*ExtendedError))) NamespaceInterface

	// Targets a room when emitting.
	To(...Room) *BroadcastOperator

	// Targets a room when emitting.
	In(...Room) *BroadcastOperator

	// Excludes a room when emitting.
	Except(...Room) *BroadcastOperator

	// Adds a new client.
	Add(*Client, any, func(*Socket))

	// Emits to all clients.
	Emit(string, ...any) error

	// Emits an event and waits for an acknowledgement from all clients.
	EmitWithAck(string, ...any) func(func([]any, error))

	// Sends a `message` event to all clients.
	Send(...any) NamespaceInterface

	// Sends a `message` event to all clients.
	Write(...any) NamespaceInterface

	// Emit a packet to other Socket.IO servers
	ServerSideEmit(string, ...any) error

	// Sends a message and expect an acknowledgement from the other Socket.IO servers of the cluster.
	ServerSideEmitWithAck(string, ...any) func(func([]any, error))

	// Gets a list of clients.
	AllSockets() (*types.Set[SocketId], error)

	// Sets the compress flag.
	Compress(bool) *BroadcastOperator

	// Sets a modifier for a subsequent event emission that the event data may be lost if the client is not ready to
	// receive messages (because of network slowness or other issues, or because they’re connected through long polling
	// and is in the middle of a request-response cycle).
	Volatile() *BroadcastOperator

	// Sets a modifier for a subsequent event emission that the event data will only be broadcast to the current node.
	Local() *BroadcastOperator

	// Adds a timeout in milliseconds for the next operation
	Timeout(time.Duration) *BroadcastOperator

	// Returns the matching socket instances
	FetchSockets() ([]*RemoteSocket, error)

	// Makes the matching socket instances join the specified rooms
	SocketsJoin(...Room)

	// Makes the matching socket instances leave the specified rooms
	SocketsLeave(...Room)

	// Makes the matching socket instances disconnect
	DisconnectSockets(bool)
}
