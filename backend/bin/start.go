package main

import (
	"net/http"

	"github.com/prettyyjnic/redisSky/backend"
	"golang.org/x/net/websocket"
)

func main() {

	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		for {
			var data backend.Message
			websocket.JSON.Receive(ws, &data)
			switch data.Operation {
			case "ping":
				var response backend.Message
				response.Operation = "pong"
				response.Data = "pong"
				go websocket.JSON.Send(ws, response)
			case "close":
				ws.Close()
				break
			case "keys":
				go backend.ScanKeys(ws, data.Data)
			}
		}
	}))
	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
