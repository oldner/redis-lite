package core

import (
	"fmt"
	"redis-lite/pkg/database"
	"strconv"
	"strings"
	"time"
)

// Eval executes a command and returns the RESP-encoded response.
func Eval(db *database.Store, args []string) []byte {
	if len(args) == 0 {
		return []byte("-ERR empty command\r\n")
	}

	errDuration := []byte("wrong duration for ttl")
	errArgLen := func(command string) []byte {
		err := fmt.Sprintf("-ERR wrong number of arguments for '%s' command\r\n", command)
		return []byte(err)
	}
	var err error

	cmd := strings.ToUpper(args[0])

	var expiry time.Duration

	switch cmd {
	case "PING":
		return []byte("+PONG\r\n")

	case "SET":
		// syntax SET key value ttl
		if len(args) < 3 {
			return errArgLen("SET")
		}
		if len(args) > 3 {
			expiry, err = time.ParseDuration(args[3])
			if err != nil {
				return errDuration
			}
		}
		// Default TTL 0
		db.Set(args[1], args[2], expiry)
		return []byte("+OK\r\n")

	case "GET":
		if len(args) != 2 {
			return errArgLen("GET")
		}
		val, found := db.Get(args[1])
		if !found {
			return []byte("$-1\r\n")
		}
		strVal, ok := val.(string)
		if !ok {
			return []byte("-ERR value is not a string\r\n")
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(strVal), strVal))

	case "DEL":
		if len(args) != 2 {
			return errArgLen("DEL")
		}
		db.Delete(args[1])
		return []byte(":1\r\n")

	case "HSET":
		// syntax: HSET key field value ttl
		if len(args) < 4 {
			return errArgLen("HSET")
		}
		if len(args) > 4 {
			expiry, err = time.ParseDuration(args[4])
			if err != nil {
				return errDuration
			}
		}
		created, err := db.HSet(args[1], args[2], args[3], expiry)
		if err != nil {
			return []byte("-ERR " + err.Error() + "\r\n")
		}
		if created {
			return []byte(":1\r\n")
		}
		return []byte(":0\r\n")

	case "HGET":
		if len(args) < 2 {
			return errArgLen("HGET")
		}
		val, found := db.HGet(args[1], args[2])
		if !found {
			return []byte("$-1\r\n")
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))

	case "LPUSH":
		if len(args) < 3 {
			return errArgLen("LPUSH")
		}
		if len(args) > 3 {
			expiry, err = time.ParseDuration(args[3])
			if err != nil {
				return errDuration
			}
		}
		count, err := db.LPush(args[1], args[2], expiry)
		if err != nil {
			return []byte("-ERR " + err.Error() + "\r\n")
		}
		return []byte(fmt.Sprintf(":%d\r\n", count))

	case "LPOP":
		if len(args) < 2 {
			return errArgLen("LPOP")
		}
		val, found := db.LPop(args[1])
		if !found {
			return []byte("$-1\r\n")
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))

	case "LRANGE":
		if len(args) < 4 {
			return errArgLen("LRANGE")
		}
		start, _ := strconv.Atoi(args[2])
		stop, _ := strconv.Atoi(args[3])

		list, found := db.LRange(args[1], start, stop)
		if !found {
			return []byte("*0\r\n")
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("*%d\r\n", len(list)))
		for _, v := range list {
			sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
		}
		return []byte(sb.String())

	case "SADD":
		if len(args) < 3 {
			return errArgLen("SADD")
		}
		added, err := db.SAdd(args[1], args[2:])
		if err != nil {
			return []byte("-ERR " + err.Error() + "\r\n")
		}
		return []byte(fmt.Sprintf(":%d\r\n", added))

	case "SMEMBERS":
		if len(args) < 2 {
			return errArgLen("SMEMBERS")
		}
		members, found := db.SMembers(args[1])
		if !found {
			return []byte("*0\r\n")
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("*%d\r\n", len(members)))
		for _, m := range members {
			sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(m), m))
		}
		return []byte(sb.String())
	case "PUBLISH":
		// syntax: PUBLISH topic message
		if len(args) < 3 {
			return errArgLen("PUBLISH")
		}
		count := db.PubSub.Publish(args[1], args[2])
		return []byte(fmt.Sprintf(":%d\r\n", count))
	default:
		return []byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", cmd))
	}
}

func IsWriteOp(cmd string) bool {
	switch strings.ToUpper(cmd) {
	case "SET", "DEL", "HSET", "LPUSH", "LPOP", "SADD":
		return true
	}
	return false
}
