package backend

import (
	"fmt"

	"reflect"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/websocket"
)

// AddKey save to redis
func AddKey(ws *websocket.Conn, data interface{}) {
	if c, _redisValue, ok := checkRedisValue(ws, data); ok {
		defer c.Close()
		switch _redisValue.T {
		case "string":
			val, ok := (_redisValue.Val).(string)
			if !ok {
				sendCmdError(ws, "val should be string")
				return
			}
			cmd := "SET " + _redisValue.Key + " " + val
			sendCmd(ws, cmd)
			r, err := c.Do("SET", _redisValue.Key, val)
			if err != nil {
				sendCmdError(ws, err.Error())
				return
			}
			sendCmdReceive(ws, r)
		case "list", "set":
			val, ok := (_redisValue.Val).([]string)
			if !ok {
				sendCmdError(ws, "val should be array of string")
				return
			}
			var method string
			if _redisValue.T == "list" {
				method = "LPUSH"
			} else {
				method = "SADD"
			}
			slice := make([]interface{}, 0, 10)
			slice = append(slice, _redisValue.Key, val)
			cmd := fmt.Sprintf("%s %v", method, slice)
			sendCmd(ws, cmd)
			r, err := c.Do(method, slice...)
			if err != nil {
				sendCmdError(ws, err.Error())
				return
			}
			sendCmdReceive(ws, r)
		case "hash":
			hset(ws, c, _redisValue)
		case "zset":
			zadd(ws, c, _redisValue)
		default:
			sendCmdError(ws, "type is not correct")
			return
		}

		message := &Message{
			Operation: "AddKey",
			Data:      "success",
		}
		websocket.JSON.Send(ws, message)
	}
}

// zadd
func zadd(ws *websocket.Conn, c redis.Conn, _redisValue redisValue) bool {
	vals, ok := (_redisValue.Val).(map[string]interface{})
	if !ok {
		sendCmdError(ws, "val should be map of string -> int64 or string -> string")
		return false
	}
	var cmd string
	for k, v := range vals {
		kind := reflect.ValueOf(v).Kind()
		if kind != reflect.Int64 && kind != reflect.Int && kind != reflect.String {
			sendCmdError(ws, "val should be map of string -> int or string -> string")
			return false
		}

		cmd = fmt.Sprintf("ZADD %s %d %s", _redisValue.Key, v, k)
		sendCmd(ws, cmd)
		r, err := c.Do("ZADD", _redisValue.Key, v, k)
		if err != nil {
			sendCmdError(ws, err.Error())
			return false
		}
		sendCmdReceive(ws, r)
	}
	return true
}

// hset redis hset
func hset(ws *websocket.Conn, c redis.Conn, _redisValue redisValue) bool {
	vals, ok := (_redisValue.Val).(map[string]interface{})
	if !ok {
		sendCmdError(ws, "val should be map of string -> int64 or string -> string")
		return false
	}
	var cmd string
	for k, v := range vals {
		kind := reflect.ValueOf(v).Kind()
		if kind != reflect.Int64 && kind != reflect.Int && kind != reflect.String {
			sendCmdError(ws, "val should be map of string -> int or string -> string")
			return false
		}

		cmd = fmt.Sprintf("HSET %s %s %s", _redisValue.Key, k, v)
		sendCmd(ws, cmd)
		r, err := c.Do("HSET", _redisValue.Key, k, v)
		if err != nil {
			sendCmdError(ws, err.Error())
			return false
		}
		sendCmdReceive(ws, r)
	}
	return true
}

// AddRow add one row 2 redis
func AddRow(ws *websocket.Conn, data interface{}) {
	if c, _redisValue, ok := checkRedisValue(ws, data); ok {
		defer c.Close()
		t, err := keyType(ws, c, _redisValue.Key)
		if err != nil {
			sendCmdError(ws, err.Error())
			return
		}
		if t != _redisValue.T {
			sendCmdError(ws, "type "+_redisValue.T+" does not match"+t)
			return
		}
		var allowType = [4]string{
			"hash",
			"zset",
			"set",
			"list",
		}
		for i := 0; i < len(allowType); i++ {
			if t == allowType[i] {
				switch t {
				case "set", "list":
					var method string
					v, ok := (_redisValue.Val).(string)
					if ok == false {
						sendCmdError(ws, "val should be string")
						return
					}
					if t == "set" {
						method = "SADD"
					} else {
						method = "LPUSH"
					}
					cmd := method + " " + _redisValue.Key + " " + v
					sendCmd(ws, cmd)
					r, err := c.Do(method, _redisValue.Key, v)
					if err != nil {
						sendCmdError(ws, err.Error())
						return
					}
					sendCmdReceive(ws, r)
				case "zset":
					if zadd(ws, c, _redisValue) == false {
						return
					}
				case "hash":
					if hset(ws, c, _redisValue) == false {
						return
					}
				}

				message := &Message{
					Operation: "AddRow",
					Data:      "Success",
				}
				websocket.JSON.Send(ws, message)
				return
			}
		}
		sendCmdError(ws, "type "+t+" does not support")
	}
}
