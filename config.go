package main

type Config struct {
	Listen        int    `default:":6479"`
	UpstreamRedis string `default:"127.0.0.1:6379"`
}

var modCommands = map[string]bool{
	"GET": true, "SET": true, "MGET": true, "MSET": true, "DEL": true,
	"EXISTS": true, "INCR": true, "DECR": true, "HGET": true, "HSET": true,
	"LPUSH": true, "RPUSH": true, "LPOP": true, "RPOP": true,
}
