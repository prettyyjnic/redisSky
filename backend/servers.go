package backend

import (
	"golang.org/x/net/websocket"
)

// redisServer 配置
type redisServer struct {
	ID     int
	Name   string
	Host   string
	Port   string
	Auth   string
	DBNums int
}

var _maxServerID int

// QueryServers 获取列表
func QueryServers(ws *websocket.Conn) {
	var message Message
	message.Data = _globalConfigs.Servers
	message.Operation = "QueryServers"
	websocket.JSON.Send(ws, message)
}

// AddServer 增加redis server
func AddServer(ws *websocket.Conn, data interface{}) {
	var server redisServer
	var err error
	var message Message
	message.Operation = "AddServer"
	server, ok := data.(redisServer)
	if ok == false {
		message.Error = err.Error()
		websocket.JSON.Send(ws, message)
		return
	}

	_maxServerID++
	server.ID = _maxServerID
	_globalConfigs.Servers = append(_globalConfigs.Servers, server)
	saveConf()
	message.Data = _globalConfigs.Servers
	websocket.JSON.Send(ws, message)
}

// UpdateServer 更新redisServer
func UpdateServer(ws *websocket.Conn, data interface{}) {
	var server redisServer
	var err error
	var message Message
	message.Operation = "UpdateServer"
	server, ok := data.(redisServer)
	if ok == false {
		message.Error = err.Error()
		websocket.JSON.Send(ws, message)
		return
	}

	for i := 0; i < len(_globalConfigs.Servers); i++ {
		if server.ID == _globalConfigs.Servers[i].ID {
			_globalConfigs.Servers[i] = server
		}
	}
	saveConf()
	message.Data = _globalConfigs.Servers
	websocket.JSON.Send(ws, message)
}
