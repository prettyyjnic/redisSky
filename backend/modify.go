package backend

import (
	"encoding/json"
	"fmt"

	"github.com/garyburd/redigo/redis"
	gosocketio "github.com/graarh/golang-socketio"
)

func lrange(conn *gosocketio.Channel, c redis.Conn, key string, start int, end int) ([]string, bool) {
	var cmd string
	cmd = fmt.Sprintf("LRANGE %s %d %d", key, start, end)
	sendCmd(conn, cmd)
	vals, err := redis.Strings(c.Do("LRANGE", key, start, end))
	if err != nil {
		sendCmdError(conn, "redis err:"+err.Error())
		return nil, false
	}
	return vals, true
}

// ModifyKey modify one key
func ModifyKey(conn *gosocketio.Channel, data interface{}) {
	if c, _redisValue, ok := checkRedisValue(conn, data); ok {
		defer c.Close()
		var key = _redisValue.Key
		t, err := keyType(conn, c, key)
		var cmd string
		if err == nil {
			switch t {
			case "none":
				sendCmdError(conn, "key: "+key+" is not exists")
			case "string":
				data, ok := (_redisValue.Val).(string)
				if !ok {
					sendCmdError(conn, "val should be string")
					return
				}
				cmd = "SET " + key + " " + data
				sendCmd(conn, cmd)
				r, err := c.Do("SET", key, data)
				if err != nil {
					sendCmdError(conn, "redis error:"+err.Error())
					return
				}
				sendCmdReceive(conn, r)
			default:
				bytes, _ := json.Marshal(_redisValue.Val)
				var _val dataStruct
				err := json.Unmarshal(bytes, &_val)
				if err != nil {
					sendCmdError(conn, "val should be dataStruct")
					return
				}
				switch t {
				case "list":
					oldVal, ok := (_val.OldVal.Val).(string)
					if ok == false {
						sendCmdError(conn, "oldval should be string")
						return
					}
					newVal, ok := (_val.NewVal.Val).(string)
					if ok == false {
						sendCmdError(conn, "newVal should be string")
						return
					}
					if vals, ok := lrange(conn, c, _redisValue.Key, _val.Index, _val.Index); ok {
						if len(vals) != 0 || len(vals) > 1 {
							sendCmdError(conn, "the index of the list is empty")
							return
						}
						var valInRedis = vals[0]
						if oldVal != valInRedis {
							sendCmdError(conn, "your value: "+valInRedis+" does not match "+oldVal)
							return
						}
						cmd = fmt.Sprintf("LSET %s %d %s", _redisValue.Key, _val.Index, newVal)
						sendCmd(conn, cmd)
						r, err := c.Do("LSET", _redisValue.Key, _val.Index, newVal)
						if err != nil {
							sendCmdError(conn, err.Error())
							return
						}
						sendCmdReceive(conn, r)
					}

				case "set":
					oldVal, ok := (_val.OldVal.Val).(string)
					if ok == false {
						sendCmdError(conn, "oldval should be string")
						return
					}
					newVal, ok := (_val.NewVal.Val).(string)
					if ok == false {
						sendCmdError(conn, "newVal should be string")
						return
					}
					srem(conn, c, _redisValue.Key, oldVal)
					cmd = fmt.Sprintf("SADD %s %s", _redisValue.Key, newVal)
					r, err := c.Do("SADD", _redisValue.Key, newVal)
					if err != nil {
						sendCmdError(conn, "val should be dataStruct")
						return
					}
					sendCmdReceive(conn, r)
				case "zset":
					oldVal, ok := (_val.OldVal.Val).(map[string]int)
					if ok == false || oldVal == nil || len(oldVal) == 0 || len(oldVal) > 1 {
						sendCmdError(conn, "oldval should be map")
						return
					}
					newVal, ok := (_val.NewVal.Val).(map[string]int)
					if ok == false || newVal == nil || len(newVal) == 0 || len(newVal) > 1 {
						sendCmdError(conn, "newVal should be map")
						return
					}
					for v := range oldVal {
						zrem(conn, c, _redisValue.Key, v)
					}
					for v, score := range newVal {
						cmd = fmt.Sprintf("ZADD %s %d %s", _redisValue.Key, score, v)
						sendCmd(conn, cmd)
						r, err := c.Do("ZADD", _redisValue.Key, score, v)
						if err != nil {
							sendCmdError(conn, "redis err"+err.Error())
							return
						}
						sendCmdReceive(conn, r)
					}
				case "hash":
					newVal, ok := (_val.NewVal.Val).(map[string]string)
					if ok == false || newVal == nil || len(newVal) == 0 || len(newVal) > 1 {
						sendCmdError(conn, "newVal should be map")
						return
					}
					for field, v := range newVal {
						cmd = fmt.Sprintf("HSET %s %s %s", _redisValue.Key, field, v)
						sendCmd(conn, cmd)
						r, err := c.Do("HSET", _redisValue.Key, field, v)
						if err != nil {
							sendCmdError(conn, "redis err"+err.Error())
							return
						}
						sendCmdReceive(conn, r)
					}
				default:
					sendCmdError(conn, "type is unknown")
				}
			}
		}
		conn.Emit("ReloadValue", _redisValue.Key)
	}
}
