package backend

import (
	"encoding/json"
	"fmt"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/websocket"
)

func lrange(ws *websocket.Conn, c redis.Conn, key string, start int, end int) ([]string, bool) {
	var cmd string
	cmd = fmt.Sprintf("LRANGE %s %d %d", key, start, end)
	sendCmd(ws, cmd)
	vals, err := redis.Strings(c.Do("LRANGE", key, start, end))
	if err != nil {
		sendCmdError(ws, "redis err:"+err.Error())
		return nil, false
	}
	return vals, true
}

// ModifyKey modify one key
func ModifyKey(ws *websocket.Conn, data interface{}) {
	if c, _redisValue, ok := checkRedisValue(ws, data); ok {
		defer c.Close()
		var key = _redisValue.Key
		t, err := keyType(ws, c, key)
		var cmd string
		if err == nil {
			switch t {
			case "none":
				sendCmdError(ws, "key: "+key+" is not exists")
			case "string":
				data, ok := (_redisValue.Val).(string)
				if !ok {
					sendCmdError(ws, "val should be string")
					return
				}
				cmd = "SET " + key + " " + data
				sendCmd(ws, cmd)
				r, err := c.Do("SET", key, data)
				if err != nil {
					sendCmdError(ws, "redis error:"+err.Error())
					return
				}
				sendCmdReceive(ws, r)
			default:
				bytes, _ := json.Marshal(_redisValue.Val)
				var _val dataStruct
				err := json.Unmarshal(bytes, &_val)
				if err != nil {
					sendCmdError(ws, "val should be dataStruct")
					return
				}
				switch t {
				case "list":
					oldVal, ok := (_val.OldVal.Val).(string)
					if ok == false {
						sendCmdError(ws, "oldval should be string")
						return
					}
					newVal, ok := (_val.NewVal.Val).(string)
					if ok == false {
						sendCmdError(ws, "newVal should be string")
						return
					}
					if vals, ok := lrange(ws, c, _redisValue.Key, _val.Index, _val.Index); ok {
						if len(vals) != 0 || len(vals) > 1 {
							sendCmdError(ws, "the index of the list is empty")
							return
						}
						var valInRedis = vals[0]
						if oldVal != valInRedis {
							sendCmdError(ws, "your value: "+valInRedis+" does not match "+oldVal)
							return
						}
						cmd = fmt.Sprintf("LSET %s %d %s", _redisValue.Key, _val.Index, newVal)
						sendCmd(ws, cmd)
						r, err := c.Do("LSET", _redisValue.Key, _val.Index, newVal)
						if err != nil {
							sendCmdError(ws, err.Error())
							return
						}
						sendCmdReceive(ws, r)
					}

				case "set":
					oldVal, ok := (_val.OldVal.Val).(string)
					if ok == false {
						sendCmdError(ws, "oldval should be string")
						return
					}
					newVal, ok := (_val.NewVal.Val).(string)
					if ok == false {
						sendCmdError(ws, "newVal should be string")
						return
					}
					srem(ws, c, _redisValue.Key, oldVal)
					cmd = fmt.Sprintf("SADD %s %s", _redisValue.Key, newVal)
					r, err := c.Do("SADD", _redisValue.Key, newVal)
					if err != nil {
						sendCmdError(ws, "val should be dataStruct")
						return
					}
					sendCmdReceive(ws, r)
				case "zset":
					oldVal, ok := (_val.OldVal.Val).(map[string]int)
					if ok == false || oldVal == nil || len(oldVal) == 0 || len(oldVal) > 1 {
						sendCmdError(ws, "oldval should be map")
						return
					}
					newVal, ok := (_val.NewVal.Val).(map[string]int)
					if ok == false || newVal == nil || len(newVal) == 0 || len(newVal) > 1 {
						sendCmdError(ws, "newVal should be map")
						return
					}
					for v := range oldVal {
						zrem(ws, c, _redisValue.Key, v)
					}
					for v, score := range newVal {
						cmd = fmt.Sprintf("ZADD %s %d %s", _redisValue.Key, score, v)
						sendCmd(ws, cmd)
						r, err := c.Do("ZADD", _redisValue.Key, score, v)
						if err != nil {
							sendCmdError(ws, "redis err"+err.Error())
							return
						}
						sendCmdReceive(ws, r)
					}
				case "hash":
					newVal, ok := (_val.NewVal.Val).(map[string]string)
					if ok == false || newVal == nil || len(newVal) == 0 || len(newVal) > 1 {
						sendCmdError(ws, "newVal should be map")
						return
					}
					for field, v := range newVal {
						cmd = fmt.Sprintf("HSET %s %s %s", _redisValue.Key, field, v)
						sendCmd(ws, cmd)
						r, err := c.Do("HSET", _redisValue.Key, field, v)
						if err != nil {
							sendCmdError(ws, "redis err"+err.Error())
							return
						}
						sendCmdReceive(ws, r)
					}
				default:
					sendCmdError(ws, "type is unknown")
				}
			}
		}
		message := &Message{
			Operation: "ModifyKey",
			Data:      "Success",
		}
		websocket.JSON.Send(ws, message)
	}
}
