package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/graarh/golang-socketio"
	"time"
)

var delKeysStorage []*delKeysStruct
var maxDelID int

type delKeysStruct struct {
	redisIns    *redis.Conn
	ID          int      `json:"id"`
	Keys        []string `json:"keys"`
	Process     float32  `json:"process"`
	ErrMsg      string   `json:"errMsg"`
	IsComplete  bool     `json:"isComplete"`
	HadTryTimes int      `json:"had_try_times"`
	ServerId    int      `json:"server_id"`
	DB          int      `json:"db"`
}

func (task *delKeysStruct) calProcess(currentKeyLen, currentKeyTotal int) {
	if isDebug {
		log.Println("calProcess", currentKeyLen, currentKeyTotal, len(task.Keys))
	}
	task.Process += float32(currentKeyLen) / float32(currentKeyTotal) / float32(len(task.Keys))
	if isDebug {
		log.Println("calProcess", task.Process)
	}
}
func (task *delKeysStruct) run() {
	go func() {
		c := *task.redisIns
		defer func() {
			task.HadTryTimes ++
			if task.ErrMsg != "" && task.HadTryTimes < 3 { // 重试
				time.Sleep(time.Second * 33)
				task.ErrMsg = ""
				task.run()
			} else {
				task.IsComplete = true
				if task.redisIns != nil {
					(*task.redisIns).Close()
				}
			}
		}()
		var err error
		if c.Err() != nil {
			c.Close()
			c, err = getRedisClient(task.ServerId, task.DB)
			if err != nil {
				log.Println("connect to redis error", err.Error())
				return
			}
			task.redisIns = &c
		}

		for index := 0; index < len(task.Keys); index++ {
			key := task.Keys[index]
			t, err := keyType(nil, c, key)
			if err != nil {
				task.Process = 0
				task.ErrMsg = err.Error()
				return
			}
			switch t {
			case "none": // 跳过不存在的key
				continue
			case "string":
				delKey(nil, c, key)
				task.calProcess(1, 1)
			case "list":
				var limits = int64(_globalConfigs.System.DelRowLimits)
				if sizes, ok := checkBigKey(nil, c, key, "list"); ok {
					// else use ltrim
					for end := sizes - 1; end >= 0; end = end - limits {
						start := end - limits
						if start < 0 {
							start = 0
						}
						_, err := c.Do("LTRIM", key, start, end)
						if err != nil {
							task.Process = 0
							task.ErrMsg = err.Error()
							return
						}
						task.calProcess(int(end-start), int(sizes))
					}
					delKey(nil, c, key)
				} else {
					task.calProcess(1, 1)
				}
			case "hash", "set", "zset":
				// var limits = int64(_globalConfigs.System.DelRowLimits)
				if _, ok := checkBigKey(nil, c, key, t); ok {
					var delMethod string
					var lenMethod string
					var _scanType scanType
					switch t {
					case "hash":
						_scanType = hashScan
						delMethod = "HDEL"
						lenMethod = "HLEN"
					case "set":
						_scanType = setScan
						delMethod = "SREM"
						lenMethod = "SCARD"
					case "zset":
						_scanType = zsetScan
						lenMethod = "ZCARD"
						delMethod = "ZREM"
					}
					var iterater int64
					var sizes int
					var err error
					sizes, err = redis.Int(c.Do(lenMethod, key))
					if err != nil {
						task.Process = 0
						task.ErrMsg = err.Error()
						return
					}
					for {
						iterater, fields := scan(nil, c, key, "", _scanType, iterater, _globalConfigs.System.RowScanLimits)
						if fields == nil {
							task.Process = 0
							task.ErrMsg = "scan key " + key + " return nil"
							return
						}
						slice := make([]interface{}, 0, _globalConfigs.System.RowScanLimits)
						slice = append(slice, key)
						for i := 0; i < len(fields); i = i + 2 {
							slice = append(slice, fields[i])
						}
						_, err = redis.Int64(c.Do(delMethod, slice...))
						slice = nil
						if err != nil {
							task.Process = 0
							task.ErrMsg = err.Error()
							return
						}
						task.calProcess(len(fields), sizes)
						if iterater == 0 {
							break
						}
					}
					delKey(nil, c, key)
				} else {
					task.calProcess(1, 1)
				}
			} // switch end
		} // for end

		task.Process = 1
	}()
}

func init() {

	delKeysStorage = make([]*delKeysStruct, 0, 10)
}

// DelKeysBg del mutil key in background
func DelKeysBg(conn *gosocketio.Channel, data json.RawMessage) {
	if operData, ok := checkOperData(conn, data); ok {
		var keys []string
		err := json.Unmarshal(operData.Data, &keys)
		if err != nil {
			sendCmdError(conn, err.Error())
			return
		}
		if len(keys) <= 0 {
			sendCmdError(conn, "keys could not be empty")
			return
		}

		c, err := getRedisClient(operData.ServerID, operData.DB)
		if err != nil {
			sendCmdError(conn, "getRedisError"+err.Error())
			return
		}
		maxDelID++
		task := &delKeysStruct{
			ServerId:   operData.ServerID,
			DB:         operData.DB,
			redisIns:   &c,
			Keys:       keys,
			ID:         maxDelID,
			IsComplete: false,
		}
		delKeysStorage = append(delKeysStorage, task)
		task.run()
		conn.Emit("AddDelTaskSuccess", 0)
	}
}

// GetDelTasksProcess get process of all tasks
func GetDelTasksProcess(conn *gosocketio.Channel, data interface{}) {
	conn.Emit("ShowDelTaskProcess", delKeysStorage)
}

// DelDeleteTask get process of all tasks
func DelDeleteTask(conn *gosocketio.Channel, id int) {
	for i := 0; i < len(delKeysStorage); i++ {
		if delKeysStorage[i].ID == id {
			if delKeysStorage[i].IsComplete == false {
				sendCmdError(conn, "del error, the task is not completed !")
				return
			}
			delKeysStorage = append(delKeysStorage[:i], delKeysStorage[i+1:]...)
			break
		}
	}
	conn.Emit("tip", &info{"success", "del success!", 2})
}

func del(conn *gosocketio.Channel, c redis.Conn, key string) bool {
	t, err := keyType(conn, c, key)
	if err != nil {
		return false
	}
	switch t {
	case "none":
		sendCmdError(conn, "key: "+key+" is not exists")
	case "string":
		delKey(conn, c, key)
	case "list":
		var limits = int64(_globalConfigs.System.DelRowLimits)
		if sizes, ok := checkBigKey(conn, c, key, "list"); ok {
			// else use ltrim
			for end := sizes - 1; end >= 0; end = end - limits {
				start := end - limits
				if start < 0 {
					start = 0
				}
				cmd := fmt.Sprintf("LTRIM %s %d %d", key, start, end)
				sendCmd(conn, cmd)
				data, err := c.Do("LTRIM", key, start, end)
				if err != nil {
					sendCmdError(conn, err.Error())
					return false
				}
				sendCmdReceive(conn, data)
			}
		}

	case "hash", "set", "zset":
		// var limits = int64(_globalConfigs.System.DelRowLimits)
		if _, ok := checkBigKey(conn, c, key, t); ok {
			var delMethod string
			var _scanType scanType
			switch t {
			case "hash":
				_scanType = hashScan
				delMethod = "HDEL"
			case "set":
				_scanType = setScan
				delMethod = "SREM"
			case "zset":
				_scanType = zsetScan
				delMethod = "ZREM"
			}
			var iterater int64

			for {
				iterater, fields := scan(conn, c, key, "", _scanType, iterater, _globalConfigs.System.RowScanLimits)
				slice := make([]interface{}, 0, _globalConfigs.System.RowScanLimits)
				slice = append(slice, key)
				for i := 0; i < len(fields); i = i + 2 {
					slice = append(slice, fields[i])
				}
				cmd := fmt.Sprintf("%s %v", delMethod, slice)
				sendCmd(conn, cmd)
				_, err := redis.Int64(c.Do(delMethod, slice...))
				slice = nil
				if err != nil {
					sendCmdError(conn, err.Error())
					return false
				}
				if iterater == 0 {
					break
				}
			}

		}

	}
	return true
	// conn.Emit("ReloadKeys", nil)
}

// DelKey del one key
func DelKey(conn *gosocketio.Channel, data json.RawMessage) {
	if c, _redisValue, ok := checkRedisValue(conn, data); ok {
		defer c.Close()
		var key = _redisValue.Key
		if del(conn, c, key) {
			conn.Emit("DelSuccess", 0)
			conn.Emit("tip", &info{"success", "del success!", 2})
		}
	}
}

func checkBigKey(conn *gosocketio.Channel, c redis.Conn, key string, t string) (int64, bool) {
	if checkLazyDel(conn, c) {
		delKey(conn, c, key)
		return 0, false
	}
	var method string
	switch t {
	case "list":
		method = "LLEN"
	case "hash":
		method = "HLEN"
	case "set":
		method = "SCARD"
	case "zset":
		method = "ZCARD"
	}
	sendCmd(conn, method+" "+key)
	sizes, err := redis.Int64(c.Do(method, key))
	if err != nil {
		sendCmdError(conn, err.Error())
		return 0, false
	}
	sendCmdReceive(conn, sizes)
	if sizes == 0 {
		sendCmdError(conn, "key is not exists")
		return 0, false
	}
	var limits = int64(_globalConfigs.System.DelRowLimits)
	if sizes <= limits { // just del it if sizes lt DelRowLimits
		delKey(conn, c, key)
		return sizes, false
	}
	return sizes, true

}

func checkLazyDel(conn *gosocketio.Channel, c redis.Conn) bool {
	infos, err := c.Do("INFO", "SERVER")
	if err != nil {
		sendCmdError(conn, err.Error())
		return false
	}
	sendCmdReceive(conn, infos)

	retBytes := bytes.Split(infos.([]byte), []byte("\r\n"))
	if err != nil {
		sendCmdError(conn, err.Error())
		return false
	}
	for i := 0; i < len(retBytes); i++ {
		info := string(retBytes[i])
		infoArr := strings.Split(info, ":")
		if len(infoArr) == 2 && infoArr[0] == "redis_version" {
			verArr := strings.Split(infoArr[0], ".")
			if len(verArr) == 3 {
				v0, _ := strconv.Atoi(verArr[0])
				if v0 > 3 {
					return true
				}
				if v0 < 3 {
					return false
				}
				if v1, _ := strconv.Atoi(verArr[1]); v1 >= 4 {
					return true
				}
				return false
			}
		}
	}
	return false
}

func delKey(conn *gosocketio.Channel, c redis.Conn, key string) {
	sendCmd(conn, "DEL "+key)
	i, err := redis.Int(c.Do("DEL", key))
	if err != nil {
		sendCmdError(conn, err.Error())
		return
	}
	sendCmdReceive(conn, i)
	if i == 0 {
		sendCmdError(conn, "key: "+key+" is not exists")
		return
	}
}

func srem(conn *gosocketio.Channel, c redis.Conn, key, val string) {
	var cmd string
	cmd = fmt.Sprintf("SREM %s %s", key, val)
	sendCmd(conn, cmd)
	r, err := c.Do("SREM", key, val)
	if err != nil {
		sendCmdError(conn, err.Error())
		return
	}
	sendCmdReceive(conn, r)
}

func zrem(conn *gosocketio.Channel, c redis.Conn, key, val string) {
	var cmd string
	cmd = fmt.Sprintf("ZREM %s %s", key, val)
	sendCmd(conn, cmd)
	r, err := c.Do("ZREM", key, val)
	if err != nil {
		sendCmdError(conn, err.Error())
		return
	}
	sendCmdReceive(conn, r)
}

func hdel(conn *gosocketio.Channel, c redis.Conn, key, field string) {
	var cmd string
	cmd = fmt.Sprintf("HDEL %s %s", key, field)
	sendCmd(conn, cmd)
	r, err := c.Do("HDEL", key, field)
	if err != nil {
		sendCmdError(conn, err.Error())
		return
	}
	sendCmdReceive(conn, r)
}

// DelRow del one row
func DelRow(conn *gosocketio.Channel, data json.RawMessage) {
	if c, _redisValue, ok := checkRedisValue(conn, data); ok {
		defer c.Close()
		var key = _redisValue.Key
		t, err := keyType(conn, c, key)
		if err == nil {
			switch t {
			case "none":
				sendCmdError(conn, "key: "+key+" is not exists")
			case "string":
				sendCmdError(conn, "string don not support this func")
			case "set", "zset", "hash":
				val, ok := (_redisValue.Val).(string)
				if !ok {
					sendCmdError(conn, "val should be string")
					return
				}
				if t == "set" {
					srem(conn, c, key, val)
				} else if t == "zset" {
					zrem(conn, c, key, val)
				} else {
					hdel(conn, c, key, val)
				}
			case "list":
				bytes, _ := json.Marshal(_redisValue.Val)
				var _val dataStruct
				err := json.Unmarshal(bytes, &_val)
				if err != nil {
					sendCmdError(conn, "val should be dataStruct")
					return
				}
				oldVal := _val.OldVal.Val
				if vals, ok := lrange(conn, c, _redisValue.Key, _val.Index, _val.Index); ok {
					if len(vals) == 0 {
						sendCmdError(conn, "the index of the list is empty")
						return
					}
					if len(vals) != 1 {
						sendCmdError(conn, "error vals")
						return
					}
					// check the field is modify already
					var valInRedis = vals[0]
					if oldVal != valInRedis {
						sendCmdError(conn, "your value: "+valInRedis+" does not match "+oldVal)
						return
					}
					removeVal := "-----TMP-----VALUE-----SHOULD-----REMOVE-----"
					cmd := fmt.Sprintf("LSET %s %d %s", _redisValue.Key, _val.Index, removeVal)
					sendCmd(conn, cmd)
					r, err := c.Do("LSET", _redisValue.Key, _val.Index, removeVal)
					if err != nil {
						sendCmdError(conn, "redis err:"+err.Error())
						return
					}
					sendCmdReceive(conn, r)

					cmd = fmt.Sprintf("LREM %s 0 %s", _redisValue.Key, removeVal)
					sendCmd(conn, cmd)
					r, err = c.Do("LREM", _redisValue.Key, 0, removeVal)
					if err != nil {
						sendCmdError(conn, "redis err:"+err.Error())
						return
					}
					sendCmdReceive(conn, r)
				}
			}

			conn.Emit("ReloadValue", 0)
		}
	}
}
