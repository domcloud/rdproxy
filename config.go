package main

type Config struct {
	Listen        string `default:":6479"`
	UpstreamRedis string `default:"127.0.0.1:6379"`
}
