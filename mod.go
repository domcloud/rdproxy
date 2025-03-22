package main

import "strings"

var modCommands = map[string]bool{
	"GET": true, "SET": true, "MGET": true, "MSET": true, "DEL": true,
	"EXISTS": true, "INCR": true, "DECR": true, "HGET": true, "HSET": true,
	"LPUSH": true, "RPUSH": true, "LPOP": true, "RPOP": true,
	"TYPE": true, "TTL": true,
}

func modSingleCommand(command, username string, args [][]byte) transformerFn {
	if modCommands[command] && len(args) > 1 {
		args[1] = append([]byte(username+":"), args[1]...)
		return nil
	}

	if command == "OBJECT" && len(args) == 3 {
		args[2] = append([]byte(username+":"), args[2]...)
		return nil
	}

	if command == "KEYS" && len(args) == 2 {
		args[1] = append([]byte(username+":"), args[1]...)
		return func(code byte, line []byte) []byte {
			if len(line) == 0 || (line[0] >= '0' && line[0] <= '9') {
				return line
			}
			return line[len(username)+1:]
		}
	}

	if command == "SCAN" {
		// Find "MATCH" argument and modify its pattern
		for i := 1; i < len(args)-1; i++ {
			if strings.ToUpper(string(args[i])) == "MATCH" {
				args[i+1] = append([]byte(username+":"), args[i+1]...)
				break
			}
		}
		return func(code byte, line []byte) []byte {
			if len(line) == 0 || (line[0] >= '0' && line[0] <= '9') {
				return line
			}
			return line[len(username)+1:]
		}
	}
	return nil
}
