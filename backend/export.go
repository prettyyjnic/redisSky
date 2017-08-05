package backend

import (
	"encoding/json"

	"github.com/garyburd/redigo/redis"
	gosocketio "github.com/graarh/golang-socketio"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type mongoConf struct {
	Addr       string `json:"addr"`
	Port       string `json:"port"`
	Database   string `json:"database"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Collection string `json:"collection"`
}

type mongoExportStruct struct {
	ID          int       `json:"id"`
	Mongodb     mongoConf `json:"mongodb"`
	Keys        []string  `json:"keys"`
	Process     float32   `json:"process"`
	ErrMsg      string    `json:"errMsg"`
	redisClient *redis.Conn
	session     *mgo.Session
}

var maxExportID int

// sendExportErrorToAll
func sendExportErrorToAll(msg string) {
	socketIOServer.BroadcastToAll("cmdErr", "Export to mongodb error: "+msg)
}

func (task *mongoExportStruct) calProcess(currentKeyLen, currentKeyTotal int) {
	task.Process += float32(currentKeyLen/currentKeyTotal) / float32(len(task.Keys))
}

func (task *mongoExportStruct) run() {
	go func() {
		defer func() {
			(*task.redisClient).Close()
			task.session.Close()
		}()
		c := *task.redisClient
		mongoClient := task.session.DB(task.Mongodb.Database).C(task.Mongodb.Collection)
		var keysNums = len(task.Keys)

		var _redisVal redisValue
		for i := 0; i < keysNums; i++ {
			key := task.Keys[i]
			_redisVal.Key = key
			t, err := redis.String(c.Do("TYPE", key))
			if err != nil {
				sendExportErrorToAll(err.Error())
				break
			}
			// remove the exists key
			_, err = mongoClient.RemoveAll(bson.M{"key": key})
			if err != nil {
				sendExportErrorToAll(err.Error())
				break
			}
			switch t {
			case "none":
				sendExportErrorToAll("key [" + key + "] is not exists")
				break
			case "string":
				str, err := redis.String(c.Do("GET", key))
				if err != nil {
					sendExportErrorToAll(err.Error())
					break
				}
				_redisVal.Val = str
				err = mongoClient.Insert(_redisVal)
				if err != nil {
					sendExportErrorToAll(err.Error())
					break
				}
				task.calProcess(1, 1)
			case "list":
				l, err := redis.Int(c.Do("LLEN", key))
				if err != nil {
					sendExportErrorToAll(err.Error())
					return
				}
				for j := 0; j < l; j += _globalConfigs.System.RowScanLimits {
					end := _globalConfigs.System.RowScanLimits
				}
				task.calProcess(1, l)
				return
			case "set":
			case "zset":
			case "hash":
			default:
				sendExportErrorToAll("keyType [" + t + "] of key [" + key + "]  is not support")
				break
			}
		}
	}()
}

var exportStorage []*mongoExportStruct

//Export2mongodb export keys to mongodb
func Export2mongodb(conn *gosocketio.Channel, data interface{}) {
	if operdata, ok := checkOperData(conn, data); ok {
		dataBytes, _ := json.Marshal(operdata)
		var exportInfo mongoExportStruct
		err := json.Unmarshal(dataBytes, &exportInfo)
		if err != nil {
			sendCmdError(conn, err.Error())
			return
		}
		session, err := mgo.DialWithInfo(&mgo.DialInfo{
			Username: exportInfo.Mongodb.Username,
			Password: exportInfo.Mongodb.Password,
			Database: exportInfo.Mongodb.Database,
			Addrs:    []string{exportInfo.Mongodb.Addr + ":" + exportInfo.Mongodb.Port},
		})
		if err != nil {
			sendCmdError(conn, err.Error())
			return
		}
		redisClient, err := getRedisClient(operdata.ServerID, operdata.DB)
		if err != nil {
			sendCmdError(conn, err.Error())
			return
		}
		maxExportID++
		exportInfo.ID = maxExportID
		exportInfo.session = session
		exportInfo.redisClient = &redisClient

		exportInfo.run()
		exportStorage = append(exportStorage, &exportInfo)
		conn.Emit("AddExportTaskSuccess", 0)
	}
}

// GetExportTasksProcess get process of all tasks
func GetExportTasksProcess(conn *gosocketio.Channel, data interface{}) {
	conn.Emit("ShowExportTaskProcess", exportStorage)
}

// DelExportTask del process task
func DelExportTask(conn *gosocketio.Channel, id int) {

	conn.Emit("tip", &info{"success", "del success!", 2})
}
