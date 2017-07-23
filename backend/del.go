package backend

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
	gosocketio "github.com/graarh/golang-socketio"
)

// DelKey del one key
func DelKey(conn *gosocketio.Channel, data interface{}) {
	if c, _redisValue, ok := checkRedisValue(conn, data); ok {
		defer c.Close()
		var key = _redisValue.Key
		t, err := keyType(conn, c, key)
		if err == nil {
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
							return
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
						iterater, fields := scan(conn, c, key, "", _scanType, iterater)
						slice := make([]interface{}, 0, _globalConfigs.System.RoconncanLimits)
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
							return
						}
						if iterater == 0 {
							break
						}
					}

				}

			}

			conn.Emit("DelKey", _redisValue.Key)
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
	info, err := redis.Strings(c.Do("INFO", "SERVER"))
	if err != nil {
		sendCmdError(conn, err.Error())
		return false
	}
	sendCmdReceive(conn, info)
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
func DelRow(conn *gosocketio.Channel, data interface{}) {
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
			case "set", "zset":
				val, ok := (_redisValue.Val).(string)
				if !ok {
					sendCmdError(conn, "val should be string")
					return
				}
				if t == "set" {
					srem(conn, c, key, val)
				} else {
					zrem(conn, c, key, val)
				}
			case "hash":
				val, ok := (_redisValue.Val).(string)
				if !ok {
					sendCmdError(conn, "val should be string")
					return
				}
				hdel(conn, c, _redisValue.Key, val)
			case "list":
				bytes, _ := json.Marshal(_redisValue.Val)
				var _val dataStruct
				err := json.Unmarshal(bytes, &_val)
				if err != nil {
					sendCmdError(conn, "val should be dataStruct")
					return
				}
				oldVal, ok := (_val.OldVal.Val).(string)
				if ok == false {
					sendCmdError(conn, "oldval should be string")
					return
				}
				if vals, ok := lrange(conn, c, _redisValue.Key, _val.Index, _val.Index); ok {
					if len(vals) != 0 || len(vals) > 1 {
						sendCmdError(conn, "the index of the list is empty")
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
			val, err := _redisValue.marshal()
			if err != nil {
				sendCmdError(conn, "marshal err:"+err.Error())
				return
			}
			conn.Emit("DelRow", val)
		}
	}
}
