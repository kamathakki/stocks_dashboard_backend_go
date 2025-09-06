package socketio

import (
	"fmt"
	"stock_automation_backend_go/shared/env"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

var io *socketio.Server
var httpProtocol = env.GetEnv[string](env.EnvKeys.PROTOCOL)
var wsProtocol []transport.Transport

type ConnState struct {
	UserId int
}

func init() {

	if httpProtocol == "http" {
		wsProtocol = []transport.Transport{websocket.Default}
	} else {
		wsProtocol = []transport.Transport{websocket.Default, polling.Default}
	}
	io = socketio.NewServer(&engineio.Options{
		Transports: wsProtocol,
	})
}

func RegisterHandlers() {

	io.OnConnect("/", func(s socketio.Conn) error {
		fmt.Println("connected:", s.ID())
		s.SetContext(&ConnState{})

		return nil
	})

	io.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice: ", msg)
		s.Emit("response", "Hello from server")
	})

	io.OnEvent("/", "uploadEvent", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		s.Emit("uploadEvent", msg)
		return "received: " + msg
	})

	io.OnError("/", func(s socketio.Conn, e error) {
		// server.Remove(s.ID())
		s.SetContext(nil)
		fmt.Println("meet error:", e)
	})

	io.OnDisconnect("/", func(s socketio.Conn, reason string) {
		// Add the Remove session id. Fixed the connection & mem leak
		s.SetContext(nil)
		fmt.Println("closed", reason)
	})
}

func GetServer() *socketio.Server {
	return io
}
