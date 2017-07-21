package backend

import (
	"encoding/json"
	"io/ioutil"

	"golang.org/x/net/websocket"
)

// _configFilePath 配置文件路径
var _configFilePath string

type systemConf struct {
	ConnectionTimeout int
	ExecutionTimeout  int
	KeyScanLimits     int
	RowScanLimits     int
	DelRowLimits      int
}

type globalConfigs struct {
	Servers []redisServer `json:"servers"`
	System  systemConf    `json:"system"`
}

var _globalConfigs globalConfigs

// Message 协议
type Message struct {
	Operation string      `json:"operation"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error"`
}

// operData 操作协议
type operData struct {
	DB       int         `json:"db"`
	ServerID int         `json:"serverid"`
	Data     interface{} `json:"data"`
}

func init() {
	_configFilePath = "./conf.json"
	conf, err := ioutil.ReadFile(_configFilePath)
	checkErr(err)

	err = json.Unmarshal(conf, &_globalConfigs)
	checkErr(err)
	_maxServerID = 0
	for i := 0; i < len(_globalConfigs.Servers); i++ {
		if _maxServerID < _globalConfigs.Servers[i].ID {
			_maxServerID = _globalConfigs.Servers[i].ID
		}
	}
}

func saveConf() error {
	data, err := json.Marshal(_globalConfigs)
	if err != nil {
		logErr(err)
		err = ioutil.WriteFile(_configFilePath, data, 0755)
	}
	return err
}

// SystemConfigs 获取系统配置信息
func SystemConfigs(ws *websocket.Conn) {
	var message Message
	message.Operation = "SystemConfigs"
	message.Data = _globalConfigs.System
	websocket.JSON.Send(ws, message)
}

// UpdateSystemConfigs 更新系统信息
func UpdateSystemConfigs(ws *websocket.Conn, data interface{}) {
	var _systemConf systemConf
	var err error
	var message Message
	message.Operation = "UpdateSystemConfigs"
	_systemConf, ok := data.(systemConf)
	if ok == false {
		message.Error = err.Error()
		websocket.JSON.Send(ws, message)
		return
	}

	_globalConfigs.System = _systemConf
	err = saveConf()
	if err != nil {
		message.Error = err.Error()
		websocket.JSON.Send(ws, message)
		return
	}
	message.Data = _globalConfigs.System
	websocket.JSON.Send(ws, message)
}
