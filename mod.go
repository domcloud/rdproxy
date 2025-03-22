package main

import "strings"

// Command modification types
const (
	ModifyFirst        = "first"
	ModifyFirstTwo     = "first_two"
	ModifyAll          = "all"
	ModifyExcludeFirst = "exclude_first"
	ModifyExcludeLast  = "exclude_last"
	ModifyExcludeOpts  = "exclude_options"
	ModifyAlternate    = "alternate"
	ModifySort         = "sort"
	ModifyEvalStyle    = "eval_style"
	ModifyScanStyle    = "scan_style"

	ReviveFirst  = "first"
	ReviveSecond = "second"
	ReviveAll    = "all"
)

// Define namespaced command rules (extracted from redis-namespace)
var namespacedCommands = map[string]string{
	"APPEND": ModifyFirst, "BITCOUNT": ModifyFirst, "BITFIELD": ModifyFirst,
	"BITOP": ModifyExcludeFirst, "BITPOS": ModifyFirst, "BLPOP": ModifyExcludeLast,
	"BRPOP": ModifyExcludeLast, "BRPOPLPUSH": ModifyExcludeLast, "BZPOPMIN": ModifyFirst,
	"BZPOPMAX": ModifyFirst, "DECR": ModifyFirst, "DECRBY": ModifyFirst, "DEL": ModifyAll,
	"DUMP": ModifyFirst, "EXISTS": ModifyAll, "EXPIRE": ModifyFirst, "EXPIREAT": ModifyFirst,
	"EXPIRETIME": ModifyFirst, "EVAL": ModifyEvalStyle, "EVALSHA": ModifyEvalStyle,
	"GET": ModifyFirst, "GETEX": ModifyFirst, "GETBIT": ModifyFirst, "GETRANGE": ModifyFirst,
	"GETSET": ModifyFirst, "HSET": ModifyFirst, "HSETNX": ModifyFirst, "HGET": ModifyFirst,
	"HINCRBY": ModifyFirst, "HINCRBYFLOAT": ModifyFirst, "HMGET": ModifyFirst,
	"HMSET": ModifyFirst, "HDEL": ModifyFirst, "HEXISTS": ModifyFirst, "HLEN": ModifyFirst,
	"HKEYS": ModifyFirst, "HSCAN": ModifyFirst, "HSCAN_EACH": ModifyFirst, "HVALS": ModifyFirst,
	"HGETALL": ModifyFirst, "INCR": ModifyFirst, "INCRBY": ModifyFirst, "INCRBYFLOAT": ModifyFirst,
	"KEYS": ModifyFirst, "LINDEX": ModifyFirst, "LINSERT": ModifyFirst, "LLEN": ModifyFirst,
	"LMOVE": ModifyFirstTwo, "LPOP": ModifyFirst, "LPOS": ModifyFirst, "LPUSH": ModifyFirst, "LPUSHX": ModifyFirst,
	"LRANGE": ModifyFirst, "LREM": ModifyFirst, "LSET": ModifyFirst, "LTRIM": ModifyFirst,
	"MAPPED_HMSET": ModifyFirst, "MAPPED_HMGET": ModifyFirst, "MAPPED_MGET": ModifyAll,
	"MAPPED_MSET": ModifyAll, "MAPPED_MSETNX": ModifyAll, "MGET": ModifyAll, "MOVE": ModifyFirst,
	"MSET": ModifyAlternate, "MSETNX": ModifyAlternate, "OBJECT": ModifyExcludeFirst,
	"PERSIST": ModifyFirst, "PEXPIRE": ModifyFirst, "PEXPIREAT": ModifyFirst, "PEXPIRETIME": ModifyFirst,
	"PSUBSCRIBE": ModifyAll, "PTTL": ModifyFirst, "PUBLISH": ModifyFirst, "PUNSUBSCRIBE": ModifyAll,
	"RENAME": ModifyAll, "RENAMENX": ModifyAll, "RESTORE": ModifyFirst, "RPOP": ModifyFirst,
	"RPOPLPUSH": ModifyAll, "RPUSH": ModifyFirst, "RPUSHX": ModifyFirst, "SADD": ModifyFirst,
	"SCARD": ModifyFirst, "SCAN": ModifyScanStyle, "SCAN_EACH": ModifyScanStyle,
	"SDIFF": ModifyAll, "SDIFFSTORE": ModifyAll, "SET": ModifyFirst, "SETBIT": ModifyFirst,
	"SETEX": ModifyFirst, "SETNX": ModifyFirst, "SETRANGE": ModifyFirst, "SINTER": ModifyAll,
	"SINTERSTORE": ModifyAll, "SISMEMBER": ModifyFirst, "SMEMBERS": ModifyFirst,
	"SMISMEMBER": ModifyFirst, "SMOVE": ModifyExcludeLast, "SORT": ModifySort, "SPOP": ModifyFirst,
	"SRANDMEMBER": ModifyFirst, "SREM": ModifyFirst, "SSCAN": ModifyFirst, "SSCAN_EACH": ModifyFirst,
	"STRLEN": ModifyFirst, "SUBSCRIBE": ModifyAll, "SUNION": ModifyAll, "SUNIONSTORE": ModifyAll,
	"TTL": ModifyFirst, "TYPE": ModifyFirst, "UNLINK": ModifyAll, "UNSUBSCRIBE": ModifyAll,
	"ZADD": ModifyFirst, "ZCARD": ModifyFirst, "ZCOUNT": ModifyFirst, "ZINCRBY": ModifyFirst,
	"ZINTERSTORE": ModifyExcludeOpts, "ZPOPMIN": ModifyFirst, "ZPOPMAX": ModifyFirst,
	"ZRANGE": ModifyFirst, "ZRANGEBYSCORE": ModifyFirst, "ZRANGEBYLEX": ModifyFirst,
	"ZRANK": ModifyFirst, "ZREM": ModifyFirst, "ZREMRANGEBYRANK": ModifyFirst,
	"ZREMRANGEBYSCORE": ModifyFirst, "ZREMRANGEBYLEX": ModifyFirst, "ZREVRANGE": ModifyFirst,
	"ZREVRANGEBYSCORE": ModifyFirst, "ZREVRANGEBYLEX": ModifyFirst, "ZREVRANK": ModifyFirst,
	"ZSCAN": ModifyFirst, "ZSCAN_EACH": ModifyFirst, "ZSCORE": ModifyFirst, "ZUNIONSTORE": ModifyExcludeOpts,
}

var reviverCommands = map[string]string{
	"BLPOP":       ReviveFirst,
	"BRPOP":       ReviveFirst,
	"KEYS":        ReviveAll,
	"MAPPED_MGET": ReviveAll,
	"SCAN":        ReviveSecond,
	"SCAN_EACH":   ReviveAll,
}

func modSingleCommand(command, username string, args [][]byte) ([][]byte, reviverFn) {
	modType, exists := namespacedCommands[strings.ToUpper(command)]
	if !exists {
		return args, nil // No modification needed
	}

	namespacePrefix := []byte(username + ":")

	switch modType {
	case ModifyFirst:
		if len(args) > 1 {
			args[1] = append(namespacePrefix, args[1]...)
		}
	case ModifyFirstTwo:
		if len(args) > 2 {
			args[1] = append(namespacePrefix, args[1]...)
			args[2] = append(namespacePrefix, args[2]...)
		}
	case ModifyAll:
		for i := 1; i < len(args); i++ {
			args[i] = append(namespacePrefix, args[i]...)
		}
	case ModifyExcludeFirst:
		for i := 2; i < len(args); i++ {
			args[i] = append(namespacePrefix, args[i]...)
		}
	case ModifyExcludeLast:
		for i := 1; i < len(args)-1; i++ {
			args[i] = append(namespacePrefix, args[i]...)
		}
	case ModifyExcludeOpts:
		if len(args) > 2 && len(args[len(args)-1]) > 0 {
			// Check if last argument is an option (e.g., weight, aggregate)
			for i := 1; i < len(args)-1; i++ {
				args[i] = append(namespacePrefix, args[i]...)
			}
		} else {
			for i := 1; i < len(args); i++ {
				args[i] = append(namespacePrefix, args[i]...)
			}
		}
	case ModifyAlternate:
		for i := 2; i < len(args); i += 2 {
			args[i] = append(namespacePrefix, args[i]...)
		}
	case ModifySort:
		if len(args) > 1 {
			args[1] = append(namespacePrefix, args[1]...)
		}
		// If second argument is not hash, modify 'by', 'store', and 'get' keys
		for i := 2; i+1 < len(args); i += 1 {
			key := strings.ToUpper(string(args[i]))
			if key == "BY" || key == "STORE" {
				i += 1
				args[i] = append(namespacePrefix, args[i]...)
			}
			if key == "LIMIT" {
				i += 2
			}
			if key == "GET" {
				i += 1
			}
		}
	case ModifyEvalStyle:
		for i := 2; i < 1+len(args)/2; i++ {
			args[i] = append(namespacePrefix, args[i]...)
		}
	case ModifyScanStyle:
		// Modify MATCH argument
		found := false
		for i := 1; i < len(args)-1; i++ {
			if strings.ToUpper(string(args[i])) == "MATCH" {
				args[i+1] = append(namespacePrefix, args[i+1]...)
				found = true
				break
			}
		}
		if !found {
			args = append(args, []byte("MATCH"), []byte(username+":*"))
		}
	}

	revType, exists := reviverCommands[strings.ToUpper(command)]
	if !exists {
		return args, nil // No reviver needed
	}

	var reviver reviverFn
	switch revType {
	case ReviveFirst:
		reviver = func(pos, depth int, line []byte) []byte {
			if depth == 1 && pos != 0 {
				return line
			}
			return line[len(namespacePrefix):]
		}
	case ReviveSecond:
		reviver = func(pos, depth int, line []byte) []byte {
			if depth == 1 && pos != 1 {
				return line
			}
			return line[len(namespacePrefix):]
		}
	case ReviveAll:
		reviver = func(pos, depth int, line []byte) []byte {
			return line[len(namespacePrefix):]
		}
	}

	return args, reviver
}
