package socketio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stock_automation_backend_go/helper"
	"time"

	socketio "github.com/doquangtan/socketio/v4"
)

// Socket.IO v4-compatible server wrapper implementing http.Handler
// and providing no-op Serve/Close to match existing gateway usage.

type SocketServer struct {
	handler http.Handler
}

func (s *SocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *SocketServer) Serve() {}

func (s *SocketServer) Close() error { return nil }

var (
	srv      *SocketServer
	ioServer broadcaster
)

type broadcaster interface {
	Emit(string, ...interface{}) error
}

func init() {
	server := socketio.New()
	ioServer = server

	server.OnConnection(func(sock *socketio.Socket) {
		fmt.Println("connected:", sock.Id)
		_, _, _, _, jobTime := helper.JobTimeEmit()
		scheduled := jobTime["scheduledForTime"].Format(time.RFC3339)
		payload := map[string]string{"scheduledForTime": scheduled}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			fmt.Println("Error encoding jobScheduledEvent:", err)
			return
		}
		fmt.Println("jobScheduledEvent:", string(jsonPayload))
		sock.Emit("jobScheduledEvent", string(jsonPayload))

		sock.On("uploadEvent", func(ev *socketio.EventPayload) {
			var msg string
			if ev != nil && len(ev.Data) > 0 {
				if s, ok := ev.Data[0].(string); ok {
					msg = s
				}
			}
			fmt.Println("uploadEvent:", msg)
			Broadcast("uploadEvent", msg)
		})

		sock.On("disconnect", func(ev *socketio.EventPayload) {
			fmt.Println("closed:", "client namespace disconnect")
		})
	})

	srv = &SocketServer{handler: server.HttpHandler()}
}

// Broadcast emits to all clients in the default namespace.
func Broadcast(event string, data any) {
	if ioServer == nil {
		return
	}
	_ = ioServer.Emit(event, data)
}

func GetServer() *SocketServer {
	return srv
}
