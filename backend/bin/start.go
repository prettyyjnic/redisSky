package main

import (
	"log"
	"net/http"

	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/prettyyjnic/redisSky/backend"
	"encoding/json"
)

func main() {
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
	backend.SetSocketIOServer(server)

	server.On("QuerySystemConfigs", func(c *gosocketio.Channel) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.QuerySystemConfigs(c)
	})

	server.On("UpdateSystemConfigs", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.UpdateSystemConfigs(c, data)
	})

	server.On("QueryServer", func(c *gosocketio.Channel, serverID int) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.QueryServer(c, serverID)
	})

	server.On("QueryServers", func(c *gosocketio.Channel) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.QueryServers(c)
	})

	server.On("DelServer", func(c *gosocketio.Channel, serverID int) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.DelServer(c, serverID)
	})

	server.On("UpdateServer", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.UpdateServer(c, data)
	})

	server.On("AddServer", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.AddServer(c, data)
	})

	server.On("ScanKeys", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.ScanKeys(c, data)
	})

	server.On("GetKey", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.GetKey(c, data)
	})

	server.On("SetTTL", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.SetTTL(c, data)
	})

	server.On("Rename", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.Rename(c, data)
	})

	server.On("DelKey", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.DelKey(c, data)
	})

	server.On("AddRow", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.AddRow(c, data)
	})

	server.On("DelRow", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.DelRow(c, data)
	})

	server.On("ModifyKey", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.ModifyKey(c, data)
	})

	server.On("AddKey", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.AddKey(c, data)
	})

	server.On("ScanRemote", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.ScanRemote(c, data)
	})

	server.On("ServerInfo", func(c *gosocketio.Channel, serverID int) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.ServerInfo(c, serverID)
	})

	server.On("GetTotalKeysNums", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.GetTotalKeysNums(c, data)
	})

	server.On("Export2mongodb", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.Export2mongodb(c, data)
	})

	server.On("GetExportTasksProcess", func(c *gosocketio.Channel, data interface{}) {
		backend.GetExportTasksProcess(c, data)
	})

	server.On("DelExportTask", func(c *gosocketio.Channel, data int) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.DelExportTask(c, data)
	})

	server.On("DelKeysBg", func(c *gosocketio.Channel, data json.RawMessage) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.DelKeysBg(c, data)
	})

	server.On("GetDelTasksProcess", func(c *gosocketio.Channel, data interface{}) {
		backend.GetDelTasksProcess(c, data)
	})

	server.On("DelDeleteTask", func(c *gosocketio.Channel, data int) {
		c.Emit("loading", 0)
		defer c.Emit("loadingComplete", 0)
		backend.DelDeleteTask(c, data)
	})

	http.HandleFunc("/socket.io/", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		server.ServeHTTP(w, r)
	})

	// server http 请求
	dir := "../../frontend"
	http.Handle("/", http.FileServer(http.Dir(dir)))

	log.Println("Serving at localhost:8089...")
	log.Fatal(http.ListenAndServe(":8089", nil))

}
