package backend

import gosocketio "github.com/graarh/golang-socketio"

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
func QueryServers(conn *gosocketio.Channel) {
	conn.Emit("ShowServers", _globalConfigs.Servers)
}

// AddServer 增加redis server
func AddServer(conn *gosocketio.Channel, data interface{}) {
	var server redisServer
	server, ok := data.(redisServer)
	if ok == false {
		sendCmdError(conn, "data should be struct of redisServer")
		return
	}
	_maxServerID++
	server.ID = _maxServerID
	_globalConfigs.Servers = append(_globalConfigs.Servers, server)
	conn.Emit("AddServerSuccess", nil)
}

// UpdateServer 更新redisServer
func UpdateServer(conn *gosocketio.Channel, data interface{}) {
	var server redisServer
	var message Message
	message.Operation = "UpdateServer"
	server, ok := data.(redisServer)
	if ok == false {
		sendCmdError(conn, "data sould be struct of redisServer")
		return
	}

	for i := 0; i < len(_globalConfigs.Servers); i++ {
		if server.ID == _globalConfigs.Servers[i].ID {
			_globalConfigs.Servers[i] = server
		}
	}
	saveConf()
	conn.Emit("UpdateServerSuccess", nil)
}
