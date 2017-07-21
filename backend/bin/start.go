package main

import (
	"log"
	"net/http"

	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

func main() {
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
	server.On("gosocketio.OnConnection", func(c *gosocketio.Channel) {
		log.Println("New client connected")
		//join them to room
		// c.Join("chat")
	})

	server.On("QueryServers", func(c *gosocketio.Channel) {
		c.Emit("QueryServers", "hello world!")
	})

	http.HandleFunc("/socket.io/", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		server.ServeHTTP(w, r)
	})
	log.Println("Serving at localhost:8090...")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
