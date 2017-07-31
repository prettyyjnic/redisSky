package main

import (
	"log"
	"net/http"

	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/prettyyjnic/redisSky/backend"
)

func main() {
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Println("New client connected")
	})

	server.On("QuerySystemConfigs", func(c *gosocketio.Channel) {
		backend.QuerySystemConfigs(c)
	})

	server.On("UpdateSystemConfigs", func(c *gosocketio.Channel, data interface{}) {
		backend.UpdateSystemConfigs(c, data)
	})

	server.On("QueryServer", func(c *gosocketio.Channel, serverID int) {
		backend.QueryServer(c, serverID)
	})

	server.On("QueryServers", func(c *gosocketio.Channel) {
		backend.QueryServers(c)
	})

	server.On("DelServer", func(c *gosocketio.Channel, serverID int) {
		backend.DelServer(c, serverID)
	})

	server.On("UpdateServer", func(c *gosocketio.Channel, data interface{}) {
		backend.UpdateServer(c, data)
	})

	server.On("AddServer", func(c *gosocketio.Channel, data interface{}) {
		backend.AddServer(c, data)
	})

	server.On("ScanKeys", func(c *gosocketio.Channel, data interface{}) {
		backend.ScanKeys(c, data)
	})

	server.On("GetKey", func(c *gosocketio.Channel, data interface{}) {
		backend.GetKey(c, data)
	})

	server.On("SetTTL", func(c *gosocketio.Channel, data interface{}) {
		backend.SetTTL(c, data)
	})

	server.On("Rename", func(c *gosocketio.Channel, data interface{}) {
		backend.Rename(c, data)
	})

	server.On("DelKey", func(c *gosocketio.Channel, data interface{}) {
		backend.DelKey(c, data)
	})

	server.On("AddRow", func(c *gosocketio.Channel, data interface{}) {
		backend.AddRow(c, data)
	})

	server.On("DelRow", func(c *gosocketio.Channel, data interface{}) {
		backend.DelRow(c, data)
	})

	server.On("ModifyKey", func(c *gosocketio.Channel, data interface{}) {
		backend.ModifyKey(c, data)
	})

	server.On("AddKey", func(c *gosocketio.Channel, data interface{}) {
		backend.AddKey(c, data)
	})

	server.On("ScanRemote", func(c *gosocketio.Channel, data interface{}) {
		backend.ScanRemote(c, data)
	})

	server.On("ServerInfo", func(c *gosocketio.Channel, serverID int) {
		backend.ServerInfo(c, serverID)
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
