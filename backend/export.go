package backend

import (
	"encoding/json"
	"strconv"

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
	mgoChan     chan *redisValue
}

var exportStorage []*mongoExportStruct
var maxExportID int

// sendExportErrorToAll
func sendExportErrorToAll(msg string) {
	socketIOServer.BroadcastToAll("cmdErr", "Export to mongodb error: "+msg)
}

func (task *mongoExportStruct) calProcess(currentKeyLen, currentKeyTotal int) {
	task.Process += float32(currentKeyLen/currentKeyTotal) / float32(len(task.Keys)*2)
}

func scanVals(c redis.Conn, key string, t scanType, iterate int) (int, []string, error) {
	var method string
	var ret []interface{}
	var keys []string
	var err error
	switch t {
	case setScan:
		method = "sscan"
	case hashScan:
		method = "hscan"
	case zsetScan:
		method = "zscan"
	}
	ret, err = redis.Values(c.Do(method, key, iterate, "COUNT", _globalConfigs.System.KeyScanLimits))
	if err != nil {
		return 0, nil, err
	}
	keys, err = redis.Strings(ret[1], nil)
	if err != nil {
		return 0, nil, err
	}
	return iterate, keys, nil
}

func (task *mongoExportStruct) run() {

	go func() {
		defer task.session.Close()
		mongoClient := task.session.DB(task.Mongodb.Database).C(task.Mongodb.Collection)
		var err error
		var _redisVal *redisValue
		for {
			select {
			case _redisVal = <-task.mgoChan:
				// remove the exists key
				_, err = mongoClient.Upsert(bson.M{"key": _redisVal.Key}, _redisVal)
				if err != nil {
					sendExportErrorToAll(err.Error())
					break
				}
				task.calProcess(1, 1)
			}
		}
	}()

	go func() {
		defer func() {
			(*task.redisClient).Close()
			close(task.mgoChan)
		}()
		c := *task.redisClient

		var keysNums = len(task.Keys)
		var _redisVal redisValue
		for i := 0; i < keysNums; i++ {
			key := task.Keys[i]
			_redisVal.Key = key
			t, err := redis.String(c.Do("TYPE", key))
			if err != nil {
				sendExportErrorToAll(err.Error())
				return
			}
			_redisVal.T = t

			switch t {
			case "none":
				sendExportErrorToAll("key [" + key + "] is not exists")
				break
			case "string":
				str, err := redis.String(c.Do("GET", key))
				if err != nil {
					sendExportErrorToAll(err.Error())
					return
				}
				_redisVal.Val = str
				_redisVal.Size = 1
			case "list":
				l, err := redis.Int(c.Do("LLEN", key))
				if err != nil {
					sendExportErrorToAll(err.Error())
					return
				}
				_redisVal.Size = l
				var tmpList []string
				tmpList = make([]string, 0, 1000)
				for j := 0; j < l; j += _globalConfigs.System.RowScanLimits {
					end := _globalConfigs.System.RowScanLimits - 1
					list, err := redis.Strings(c.Do("LRANGE", key, j, end))
					if err != nil {
						sendExportErrorToAll(err.Error())
						return
					}
					tmpList = append(tmpList, list...)
					task.calProcess(len(list), l)
				}
				_redisVal.Val = tmpList
			case "set":
				l, err := redis.Int(c.Do("SCARD", key))
				if err != nil {
					sendExportErrorToAll(err.Error())
					return
				}
				_redisVal.Size = l
				var tmpList []string
				tmpList = make([]string, 0, 1000)
				var iter int
				for {
					iter, vals, err := scanVals(c, key, setScan, iter)
					if err != nil {
						sendExportErrorToAll(err.Error())
						return
					}
					task.calProcess(len(vals), l)
					tmpList = append(tmpList, vals...)
					if iter == 0 {
						break
					}
				}
				_redisVal.Val = tmpList
			case "zset":
				l, err := redis.Int(c.Do("ZCOUNT", key, "-inf", "+inf"))
				if err != nil {
					sendExportErrorToAll(err.Error())
					return
				}
				_redisVal.Size = l
				var tmpMap map[string]float64
				tmpMap = make(map[string]float64)
				var iter int
				for {
					iter, vals, err := scanVals(c, key, zsetScan, iter)
					if err != nil {
						sendExportErrorToAll(err.Error())
						return
					}
					task.calProcess(len(vals), l*2)
					for j := 0; j < len(vals); j += 2 {
						tmpFloat64, _ := strconv.ParseFloat(vals[j+1], 64)
						tmpMap[vals[j]] = tmpFloat64
					}
					if iter == 0 {
						break
					}
				}
				_redisVal.Val = tmpMap
			case "hash":
				l, err := redis.Int(c.Do("HLEN", key))
				if err != nil {
					sendExportErrorToAll(err.Error())
					return
				}
				_redisVal.Size = l
				var tmpMap map[string]string
				tmpMap = make(map[string]string)
				var iter int
				for {
					iter, vals, err := scanVals(c, key, zsetScan, iter)
					if err != nil {
						sendExportErrorToAll(err.Error())
						return
					}
					task.calProcess(len(vals), l*2)
					for j := 0; j < len(vals); j += 2 {
						tmpMap[vals[j]] = vals[j+1]
					}
					if iter == 0 {
						break
					}
				}
				_redisVal.Val = tmpMap
			default:
				sendExportErrorToAll("keyType [" + t + "] of key [" + key + "]  is not support")
				break
			}

		}

	}()
}

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
		exportInfo.mgoChan = make(chan *redisValue, 1)
		exportStorage = append(exportStorage, &exportInfo)
		exportInfo.run()
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
