package backend

import (
	"encoding/json"
	"io/ioutil"

	gosocketio "github.com/graarh/golang-socketio"
)

// _configFilePath 配置文件路径
var _configFilePath string

type systemConf struct {
	ConnectionTimeout int `json:"connectionTimeout"`
	ExecutionTimeout  int `json:"executionTimeout"`
	KeyScanLimits     int `json:"keyScanLimits"`
	RowScanLimits     int `json:"rowScanLimits"`
	DelRowLimits      int `json:"delRowLimits"`
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

func (message Message) marshal() ([]byte, error) {
	return json.Marshal(message)
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

// QuerySystemConfigs 获取系统配置信息
func QuerySystemConfigs(conn *gosocketio.Channel) {
	conn.Emit("LoadSystemConfigs", _globalConfigs.System)
}

// UpdateSystemConfigs 更新系统信息
func UpdateSystemConfigs(conn *gosocketio.Channel, data interface{}) {
	var _systemConf systemConf
	var err error
	bytes, _ := json.Marshal(data)
	err = json.Unmarshal(bytes, &_systemConf)
	if err != nil {
		sendCmdError(conn, "data sould be struct of systemConf!")
		return
	}

	_globalConfigs.System = _systemConf
	err = saveConf()
	if err != nil {
		sendCmdError(conn, err.Error())
		return
	}
	conn.Emit("LoadSystemConfigs", _globalConfigs.System)
}
