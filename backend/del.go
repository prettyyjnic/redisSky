package backend

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/websocket"
)

// DelKey del one key
func DelKey(ws *websocket.Conn, data interface{}) {
	if c, _redisValue, ok := checkRedisValue(ws, data); ok {
		defer c.Close()
		var key = _redisValue.Key
		t, err := keyType(ws, c, key)
		if err == nil {
			switch t {
			case "none":
				sendCmdError(ws, "key: "+key+" is not exists")
			case "string":
				delKey(ws, c, key)

			case "list":
				var limits = int64(_globalConfigs.System.DelRowLimits)
				if sizes, ok := checkBigKey(ws, c, key, "list"); ok {
					// else use ltrim
					for end := sizes - 1; end >= 0; end = end - limits {
						start := end - limits
						if start < 0 {
							start = 0
						}
						cmd := fmt.Sprintf("LTRIM %s %d %d", key, start, end)
						sendCmd(ws, cmd)
						data, err := c.Do("LTRIM", key, start, end)
						if err != nil {
							sendCmdError(ws, err.Error())
							return
						}
						sendCmdReceive(ws, data)
					}
				}

			case "hash", "set", "zset":
				// var limits = int64(_globalConfigs.System.DelRowLimits)
				if _, ok := checkBigKey(ws, c, key, t); ok {
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
						iterater, fields := scan(ws, c, key, "", _scanType, iterater)
						slice := make([]interface{}, 0, _globalConfigs.System.RowScanLimits)
						slice = append(slice, key)
						for i := 0; i < len(fields); i = i + 2 {
							slice = append(slice, fields[i])
						}
						cmd := fmt.Sprintf("%s %v", delMethod, slice)
						sendCmd(ws, cmd)
						_, err := redis.Int64(c.Do(delMethod, slice...))
						slice = nil
						if err != nil {
							sendCmdError(ws, err.Error())
							return
						}
						if iterater == 0 {
							break
						}
					}

				}

			}
			message := &Message{
				Operation: "DelKey",
				Data:      "OK",
			}
			websocket.JSON.Send(ws, message)
		}
	}
}

func checkBigKey(ws *websocket.Conn, c redis.Conn, key string, t string) (int64, bool) {
	if checkLazyDel(ws, c) {
		delKey(ws, c, key)
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
	sendCmd(ws, method+" "+key)
	sizes, err := redis.Int64(c.Do(method, key))
	if err != nil {
		sendCmdError(ws, err.Error())
		return 0, false
	}
	sendCmdReceive(ws, sizes)
	if sizes == 0 {
		sendCmdError(ws, "key is not exists")
		return 0, false
	}
	var limits = int64(_globalConfigs.System.DelRowLimits)
	if sizes <= limits { // just del it if sizes lt DelRowLimits
		delKey(ws, c, key)
		return sizes, false
	}
	return sizes, true

}

func checkLazyDel(ws *websocket.Conn, c redis.Conn) bool {
	info, err := redis.Strings(c.Do("INFO", "SERVER"))
	if err != nil {
		sendCmdError(ws, err.Error())
		return false
	}
	sendCmdReceive(ws, info)
	for i := 0; i < len(info); i++ {
		infoArr := strings.Split(info[i], ":")
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

func delKey(ws *websocket.Conn, c redis.Conn, key string) {
	sendCmd(ws, "DEL "+key)
	i, err := redis.Int(c.Do("DEL", key))
	if err != nil {
		sendCmdError(ws, err.Error())
		return
	}
	sendCmdReceive(ws, i)
	if i == 0 {
		sendCmdError(ws, "key: "+key+" is not exists")
		return
	}
}

func srem(ws *websocket.Conn, c redis.Conn, key, val string) {
	var cmd string
	cmd = fmt.Sprintf("SREM %s %s", key, val)
	sendCmd(ws, cmd)
	r, err := c.Do("SREM", key, val)
	if err != nil {
		sendCmdError(ws, err.Error())
		return
	}
	sendCmdReceive(ws, r)
}

func zrem(ws *websocket.Conn, c redis.Conn, key, val string) {
	var cmd string
	cmd = fmt.Sprintf("ZREM %s %s", key, val)
	sendCmd(ws, cmd)
	r, err := c.Do("ZREM", key, val)
	if err != nil {
		sendCmdError(ws, err.Error())
		return
	}
	sendCmdReceive(ws, r)
}

func hdel(ws *websocket.Conn, c redis.Conn, key, field string) {
	var cmd string
	cmd = fmt.Sprintf("HDEL %s %s", key, field)
	sendCmd(ws, cmd)
	r, err := c.Do("HDEL", key, field)
	if err != nil {
		sendCmdError(ws, err.Error())
		return
	}
	sendCmdReceive(ws, r)
}

// DelRow del one row
func DelRow(ws *websocket.Conn, data interface{}) {
	if c, _redisValue, ok := checkRedisValue(ws, data); ok {
		defer c.Close()
		var key = _redisValue.Key
		t, err := keyType(ws, c, key)
		if err == nil {
			switch t {
			case "none":
				sendCmdError(ws, "key: "+key+" is not exists")
			case "string":
				sendCmdError(ws, "string don not support this func")
			case "set", "zset":
				val, ok := (_redisValue.Val).(string)
				if !ok {
					sendCmdError(ws, "val should be string")
					return
				}
				if t == "set" {
					srem(ws, c, key, val)
				} else {
					zrem(ws, c, key, val)
				}
			case "hash":
				val, ok := (_redisValue.Val).(string)
				if !ok {
					sendCmdError(ws, "val should be string")
					return
				}
				hdel(ws, c, _redisValue.Key, val)
			case "list":
				bytes, _ := json.Marshal(_redisValue.Val)
				var _val dataStruct
				err := json.Unmarshal(bytes, &_val)
				if err != nil {
					sendCmdError(ws, "val should be dataStruct")
					return
				}
				oldVal, ok := (_val.OldVal.Val).(string)
				if ok == false {
					sendCmdError(ws, "oldval should be string")
					return
				}
				if vals, ok := lrange(ws, c, _redisValue.Key, _val.Index, _val.Index); ok {
					if len(vals) != 0 || len(vals) > 1 {
						sendCmdError(ws, "the index of the list is empty")
						return
					}
					// check the field is modify already
					var valInRedis = vals[0]
					if oldVal != valInRedis {
						sendCmdError(ws, "your value: "+valInRedis+" does not match "+oldVal)
						return
					}
					removeVal := "-----TMP-----VALUE-----SHOULD-----REMOVE-----"
					cmd := fmt.Sprintf("LSET %s %d %s", _redisValue.Key, _val.Index, removeVal)
					sendCmd(ws, cmd)
					r, err := c.Do("LSET", _redisValue.Key, _val.Index, removeVal)
					if err != nil {
						sendCmdError(ws, "redis err:"+err.Error())
						return
					}
					sendCmdReceive(ws, r)

					cmd = fmt.Sprintf("LREM %s 0 %s", _redisValue.Key, removeVal)
					sendCmd(ws, cmd)
					r, err = c.Do("LREM", _redisValue.Key, 0, removeVal)
					if err != nil {
						sendCmdError(ws, "redis err:"+err.Error())
						return
					}
					sendCmdReceive(ws, r)
				}
			}

			message := &Message{
				Operation: "DelRow",
				Data:      "Success",
			}
			websocket.JSON.Send(ws, message)
		}
	}
}
