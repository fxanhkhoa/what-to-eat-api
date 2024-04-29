package transports

import (
	"io"

	ws "github.com/gorilla/websocket"
	"github.com/zishang520/engine.io-go-parser/packet"
	_types "github.com/zishang520/engine.io-go-parser/types"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/types"
)

var ws_log = log.NewLog("engine:ws")

type websocket struct {
	*transport

	socket *types.WebSocketConn
}

// WebSocket transport
func NewWebSocket(ctx *types.HttpContext) *websocket {
	w := &websocket{}
	return w.New(ctx)
}

func (w *websocket) New(ctx *types.HttpContext) *websocket {
	w.transport = &transport{}

	// Advertise framing support.
	w.supportsFraming = true

	// Advertise upgrade support.
	w.handlesUpgrades = true

	// Transport name
	w.name = "websocket"

	w.transport.New(ctx)

	w.socket = ctx.Websocket
	w.SetWritable(true)
	w.perMessageDeflate = nil

	w.doClose = w.WebSocketDoClose
	w.send = w.WebSocketSend

	go w._init()

	w.socket.On("error", func(errors ...any) {
		w.OnError("websocket error", errors[0].(error))
	})
	w.socket.On("close", func(...any) {
		w.OnClose()
	})

	return w
}

func (w *websocket) _init() {
	for {
		mt, message, err := w.socket.NextReader()
		if err != nil {
			if ws.IsUnexpectedCloseError(err) {
				w.OnClose()
			} else {
				w.OnError("Error reading data", err)
			}
			break
		}

		switch mt {
		case ws.BinaryMessage:
			read := _types.NewBytesBuffer(nil)
			if _, err := read.ReadFrom(message); err != nil {
				w.OnError("Error reading data", err)
			} else {
				w.WebSocketOnData(read)
			}
		case ws.TextMessage:
			read := _types.NewStringBuffer(nil)
			if _, err := read.ReadFrom(message); err != nil {
				w.OnError("Error reading data", err)
			} else {
				w.WebSocketOnData(read)
			}
		case ws.CloseMessage:
			w.OnClose()
			if c, ok := message.(io.Closer); ok {
				c.Close()
			}
			break
		case ws.PingMessage:
		case ws.PongMessage:
		}
		if c, ok := message.(io.Closer); ok {
			c.Close()
		}
	}
}

func (w *websocket) WebSocketOnData(data _types.BufferInterface) {
	ws_log.Debug(`websocket received "%s"`, data)
	w.TransportOnData(data)
}

// Writes a packet payload.
func (w *websocket) WebSocketSend(packets []*packet.Packet) {
	w.SetWritable(false)
	defer func() {
		w.SetWritable(true)
		w.Emit("drain")
	}()

	w.musend.Lock()
	defer w.musend.Unlock()
	for _, packet := range packets {
		w._send(packet)
	}
}

func (w *websocket) _send(packet *packet.Packet) {
	var data _types.BufferInterface

	if packet.WsPreEncoded != nil {
		data = packet.WsPreEncoded
	} else {
		var err error
		data, err = w.parser.EncodePacket(packet, w.supportsBinary)
		if err != nil {
			ws_log.Debug(`Send Error "%s"`, err)
			return
		}
	}

	// always creates a new object since ws modifies it
	compress := false
	if packet.Options != nil {
		compress = packet.Options.Compress
	}
	if w.perMessageDeflate != nil {
		if data.Len() < w.perMessageDeflate.Threshold {
			compress = false
		}
	}
	ws_log.Debug(`writing "%s"`, data)

	w.write(data, compress)
}

func (w *websocket) write(data _types.BufferInterface, compress bool) {
	w.socket.EnableWriteCompression(compress)
	mt := ws.BinaryMessage
	if _, ok := data.(*_types.StringBuffer); ok {
		mt = ws.TextMessage
	}
	write, err := w.socket.NextWriter(mt)
	if err != nil {
		w.OnError("write error", err)
		return
	}
	defer func() {
		if err := write.Close(); err != nil {
			w.OnError("write error", err)
			return
		}
	}()
	if _, err := io.Copy(write, data); err != nil {
		w.OnError("write error", err)
		return
	}
}

// Closes the transport.
func (w *websocket) WebSocketDoClose(fn ...types.Callable) {
	ws_log.Debug(`closing`)
	if len(fn) > 0 {
		(fn[0])()
	}
	w.socket.Close()
}
