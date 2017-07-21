package backend

import (
	"errors"

	"golang.org/x/net/websocket"

	"time"

	"strconv"

	"reflect"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

var redisClients map[int]*redis.Pool

type scanType int

var pool *redis.Pool

const (
	hashScan scanType = iota
	zsetScan
	setScan
	keyScan
)

type dataStruct struct {
	Index  int        `json:"index"` // list 用
	OldVal redisValue `json:"oldVal"`
	NewVal redisValue `json:"newVal"`
}

func init() {
	redisClients = make(map[int]*redis.Pool)
}

func getRedisClient(serverID int, db int) (redis.Conn, error) {

	var ok bool

	pool, ok = redisClients[serverID]
	if ok == false {
		pool = redis.NewPool(func() (redis.Conn, error) {
			for i := 0; i < len(_globalConfigs.Servers); i++ {
				if serverID == _globalConfigs.Servers[i].ID {
					c, err := redis.DialTimeout("tcp", _globalConfigs.Servers[i].Host+":"+_globalConfigs.Servers[i].Port, time.Duration(_globalConfigs.System.ConnectionTimeout)*time.Second, time.Duration(_globalConfigs.System.ExecutionTimeout)*time.Second, time.Duration(_globalConfigs.System.ExecutionTimeout)*time.Second)
					if err != nil {
						return nil, errors.New("redis server dial error" + err.Error())
					}
					if _globalConfigs.Servers[i].Auth != "" {
						if _, err := c.Do("AUTH", _globalConfigs.Servers[i].Auth); err != nil {
							c.Close()
							return nil, err
						}
					}
					return c, nil
				}
			}
			return nil, errors.New("redis server id is out of range")
		}, 2)
	}
	if pool == nil {
		return nil, errors.New("redis server id is out of range")
	}
	c := pool.Get()
	_, err := c.Do("SELECT", db)
	if err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func closeClient(serverID int) error {
	if client, ok := redisClients[serverID]; ok {
		return client.Close()
	}
	return nil
}

// type redisValue
type redisValue struct {
	Key string      `json:"key"`
	T   string      `json:"t"`
	TTL int64       `json:"ttl"`
	Val interface{} `json:"val"`
}

// checkOperData 检查协议
func checkOperData(ws *websocket.Conn, data interface{}) (operData, bool) {
	var info operData
	if reflect.ValueOf(data).Kind() != reflect.Map {
		sendCmdError(ws, "proto error ")
		return info, false
	}
	var err error
	var bytes []byte
	bytes, err = json.Marshal(data)
	if err != nil {
		sendCmdError(ws, err.Error())
		return info, false
	}
	err = json.Unmarshal(bytes, &info)
	if err != nil {
		sendCmdError(ws, err.Error())
		return info, false
	}
	return info, true
}

// sendCmd
func sendCmd(ws *websocket.Conn, cmd string) {
	var message Message
	message.Operation = "cmd"
	message.Data = cmd
	websocket.JSON.Send(ws, message)
}

// sendRedisErr
func sendCmdError(ws *websocket.Conn, cmd string) {
	var message Message
	message.Operation = "cmd"
	message.Error = cmd
	websocket.JSON.Send(ws, message)
}

// sendRedisReceive
func sendCmdReceive(ws *websocket.Conn, data interface{}) {
	var info string
	v := reflect.ValueOf(data)
	k := v.Kind()
	switch k {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Int64:
		info = strconv.FormatInt(v.Int(), 10)
	case reflect.Float64, reflect.Float32:
		info = strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Array, reflect.Map, reflect.Slice:
		info = "Array: " + strconv.Itoa(v.Len())
	case reflect.String:
		info = "String: " + v.String()
	default:
		info = "Unknown: " + k.String()
	}

	var message Message
	message.Operation = "cmd"
	message.Data = "Receive: " + info
	websocket.JSON.Send(ws, message)
}

func checkRedisValue(ws *websocket.Conn, data interface{}) (c redis.Conn, _redisValue redisValue, b bool) {
	if info, ok := checkOperData(ws, data); ok {
		bytes, _ := json.Marshal(data)
		var err error
		err = json.Unmarshal(bytes, &_redisValue)
		if err != nil {
			sendCmdError(ws, "Unmarshal error "+err.Error())
			return c, _redisValue, false
		}
		c, err = getRedisClient(info.ServerID, info.DB)
		if err != nil {
			sendCmdError(ws, "redis error: "+err.Error())
			return c, _redisValue, false
		}
		return c, _redisValue, true
	}
	return c, _redisValue, false
}
