package backend

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

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
	ID           int       `json:"id"`
	Mongodb      mongoConf `json:"mongodb"`
	Keys         []string  `json:"keys"`
	Process      float32   `json:"process"`
	ErrMsg       string    `json:"errMsg"`
	Task         string    `json:"task"`
	processMutex *sync.RWMutex
	redisClient  *redis.Conn
	session      *mgo.Session
	mgoChan      chan redisValue
	waitGroup    *sync.WaitGroup
	errorChan    chan error
}

var exportStorage []*mongoExportStruct
var maxExportID int

// sendExportErrorToAll
func sendExportErrorToAll(msg string) {
	if isDebug {
		logErr(errors.New(msg))
	}
	socketIOServer.BroadcastToAll("cmdErr", "Export to mongodb error: "+msg)
}

func (task *mongoExportStruct) calProcess(currentKeyLen, currentKeyTotal int) {
	if isDebug {
		log.Println("calProcess", currentKeyLen, currentKeyTotal, len(task.Keys))
	}
	task.processMutex.Lock()
	task.Process += float32(currentKeyLen) / float32(currentKeyTotal) / float32(len(task.Keys)*2)
	task.processMutex.Unlock()
	if isDebug {
		log.Println("calProcess", task.Process)
	}
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
	iterate, err = redis.Int(ret[0], nil)
	if err != nil {
		return 0, nil, err
	}
	return iterate, keys, nil
}

func (task *mongoExportStruct) run() {
	task.waitGroup.Add(2)
	go func() {
		defer func() {
			task.waitGroup.Done()
			task.session.Close()
			if isDebug {
				log.Println("close mongo goroutine ")
			}
		}()
		mongoClient := task.session.DB(task.Mongodb.Database).C(task.Mongodb.Collection)
		var err error
		var _redisVal redisValue
		var isClose bool
		for {
			select {
			case err = <-task.errorChan:
				if err != nil {
					task.ErrMsg = err.Error()
					if isDebug {
						logErr(err)
					}
					return
				}
			case _redisVal, isClose = <-task.mgoChan:
				if _redisVal.Val != nil {
					// upsert key
					if isDebug {
						log.Println("upsert key!" + _redisVal.Key)
					}
					// remove the exists key
					_, err = mongoClient.Upsert(bson.M{"key": _redisVal.Key}, _redisVal)

					if err != nil {
						if isDebug {
							logErr(err)
						}
						task.errorChan <- err
						return
					}
					task.calProcess(1, 1)
				}
				if _redisVal.Val == nil && !isClose {
					return
				}
			}
		}
	}()

	go func() {
		defer func() {
			task.waitGroup.Done()
			(*task.redisClient).Close()
			close(task.mgoChan)
		}()
		c := *task.redisClient

		var keysNums = len(task.Keys)
		var _redisVal redisValue
		var err error

		for i := 0; i < keysNums; i++ {
			select {
			case err = <-task.errorChan:
				if err != nil {
					task.ErrMsg = err.Error()
					if isDebug {
						logErr(err)
					}
					return
				}
			default:
				key := task.Keys[i]
				t, err := redis.String(c.Do("TYPE", key))
				if err != nil {
					task.errorChan <- err
					return
				}
				_redisVal.Key = key
				_redisVal.T = t
				_redisVal.Val = nil
				switch t {
				case "none":
					task.errorChan <- errors.New("redis: key [" + key + "] is not exists")
					return
				case "string":
					str, err := redis.String(c.Do("GET", key))
					if err != nil {
						if isDebug {
							logErr(err)
						}
						task.errorChan <- err
						return
					}
					_redisVal.Val = str
					_redisVal.Size = 1
				case "list":
					l, err := redis.Int(c.Do("LLEN", key))
					if err != nil {
						if isDebug {
							logErr(err)
						}
						task.errorChan <- err
						return
					}
					_redisVal.Size = l
					var tmpList []string
					tmpList = make([]string, 0, 1000)
					for j := 0; j < l; j += _globalConfigs.System.RowScanLimits {
						end := _globalConfigs.System.RowScanLimits - 1
						list, err := redis.Strings(c.Do("LRANGE", key, j, end))
						if err != nil {
							if isDebug {
								logErr(err)
							}
							task.errorChan <- err
							return
						}
						tmpList = append(tmpList, list...)
						task.calProcess(len(list), l)
					}
					_redisVal.Val = tmpList
				case "set":
					l, err := redis.Int(c.Do("SCARD", key))
					if err != nil {
						task.errorChan <- err
						return
					}
					_redisVal.Size = l
					var tmpList []string
					tmpList = make([]string, 0, 1000)
					var iter int
					var vals []string
					for {
						iter, vals, err = scanVals(c, key, setScan, iter)
						if err != nil {
							if isDebug {
								logErr(err)
							}
							task.errorChan <- err
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
						task.errorChan <- err
						return
					}
					_redisVal.Size = l
					var tmpMap map[string]float64
					tmpMap = make(map[string]float64)
					var iter int
					var vals []string
					for {
						iter, vals, err = scanVals(c, key, zsetScan, iter)
						if err != nil {
							if isDebug {
								logErr(err)
							}
							task.errorChan <- err
							return
						}
						task.calProcess(len(vals)/2, l)
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
						if isDebug {
							logErr(err)
						}
						task.errorChan <- err
						return
					}
					_redisVal.Size = l
					var tmpMap map[string]string
					tmpMap = make(map[string]string)
					var iter int
					var vals []string
					for {
						iter, vals, err = scanVals(c, key, hashScan, iter)
						if err != nil {
							if isDebug {
								logErr(err)
							}
							task.errorChan <- err
							return
						}
						if isDebug {
							log.Println("iter", iter)
						}
						task.calProcess(len(vals)/2, l)
						for j := 0; j < len(vals); j += 2 {
							tmpMap[vals[j]] = vals[j+1]
						}
						if iter == 0 {
							break
						}
					}
					_redisVal.Val = tmpMap
				default:
					task.errorChan <- errors.New("redis: keyType [" + t + "] of key [" + key + "]  is not support")
					return
				}

				task.mgoChan <- _redisVal
			}
		}

	}()
	task.waitGroup.Wait()
	task.Process = 1
	select {
	case err := <-task.errorChan:
		if err != nil {
			task.ErrMsg = err.Error()
		}
	}
	close(task.errorChan)
}

//Export2mongodb export keys to mongodb
func Export2mongodb(conn *gosocketio.Channel, data interface{}) {
	if operdata, ok := checkOperData(conn, data); ok {
		dataBytes, _ := json.Marshal(operdata.Data)
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
		if exportInfo.Task == "" {
			exportInfo.Task = time.Now().Format("2006-01-02 15:04:05")
		}
		exportInfo.session = session
		exportInfo.redisClient = &redisClient
		exportInfo.mgoChan = make(chan redisValue, 5)
		exportInfo.errorChan = make(chan error, 1)
		exportInfo.processMutex = new(sync.RWMutex)
		exportInfo.waitGroup = new(sync.WaitGroup)
		exportStorage = append(exportStorage, &exportInfo)
		go exportInfo.run()
		conn.Emit("AddExportTaskSuccess", 0)
	}
}

// GetExportTasksProcess get process of all tasks
func GetExportTasksProcess(conn *gosocketio.Channel, data interface{}) {
	conn.Emit("ShowExportTaskProcess", exportStorage)
}

// DelExportTask del process task
func DelExportTask(conn *gosocketio.Channel, id int) {
	for i := 0; i < len(exportStorage); i++ {
		if exportStorage[i].ID == id {
			exportStorage = append(exportStorage[:i], exportStorage[i+1:]...)
			break
		}
	}
	conn.Emit("tip", &info{"success", "del success!", 2})
}
