package main

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"strings"

	"github.com/tidwall/redcon"
)

type Handler struct {
	config  Config
	netMode string
}

type HandlerContext struct {
	upstreamConn net.Conn
	username     string
	detached     bool
}

func NewHandler(config Config) *Handler {
	netMode := "tcp"
	if strings.HasPrefix(config.UpstreamRedis, "/") || strings.HasPrefix(config.UpstreamRedis, "@") {
		netMode = "unix"
	}
	return &Handler{
		config:  config,
		netMode: netMode,
	}
}

func (m *Handler) ServeRESP(conn redcon.Conn, cmd redcon.Command) {
	context, ok := conn.Context().(HandlerContext)
	if !ok || context.upstreamConn == nil {
		conn.Close()
		return
	}

	upConn := context.upstreamConn
	command := strings.ToUpper(string(cmd.Args[0]))
	newArgs, reviver := modSingleCommand(command, context.username, cmd.Args)
	cmd.Args = newArgs

	// Construct RESP command & send to redis
	request := buildRESPCommand(cmd.Args)
	_, err := upConn.Write(request)
	if err != nil {
		log.Printf("Failed to send command: %v", err)
		conn.Close()
		return
	}

	// Read response from Redis
	var b bytes.Buffer
	reader := newRespReader(bufio.NewReader(upConn), &b, reviver)
	if err = reader.ReadReply(); err != nil {
		log.Printf("Failed to read response: %v", err)
		conn.Close()
		return
	}
	response := b.Bytes()

	if len(response) > 0 && response[0] != '-' {
		if command == "AUTH" && len(cmd.Args) == 3 {
			context.username = string(cmd.Args[1])
			conn.SetContext(context)
		} else if command == "SUBSCRIBE" || command == "PSUBSRIBE" {
			// this will detach connection from this loop
			context.detached = true
			conn.SetContext(context)
			dconn := conn.Detach()
			go m.subscriptionLoop(dconn, upConn)
		}
	}

	// Send response back to client
	conn.WriteRaw(response)
}

func (m *Handler) AcceptConn(conn redcon.Conn) bool {
	redisConn, err := net.Dial(m.netMode, m.config.UpstreamRedis)
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
		return false
	}

	conn.SetContext(HandlerContext{
		upstreamConn: redisConn,
		username:     "default", // TODO: Fetch username dynamically
	})
	return true
}

func (m *Handler) ClosedConn(conn redcon.Conn, err error) {
	context, ok := conn.Context().(HandlerContext)
	if ok && context.upstreamConn != nil && !context.detached {
		context.upstreamConn.Close()
		context.upstreamConn = nil
	}
}

func (m *Handler) subscriptionLoop(conn redcon.DetachedConn, upConn net.Conn) {
	defer conn.Close()
	context, ok := conn.Context().(HandlerContext)
	if !ok || context.upstreamConn == nil {
		conn.Close()
		return
	}

	go (func() {
		for {
			cmd, err := conn.ReadCommand()
			if err != nil {
				log.Printf("Failed to read command: %v", err)
				conn.Close()
				return
			}
			command := strings.ToUpper(string(cmd.Args[0]))
			newArgs, _ := modSingleCommand(command, context.username, cmd.Args)
			cmd.Args = newArgs

			// Construct RESP command & send to redis
			request := buildRESPCommand(cmd.Args)
			_, err = upConn.Write(request)
			if err != nil {
				log.Printf("Failed to send command: %v", err)
				conn.Close()
				return
			}
		}
	})()

	for {
		conn.Flush()

		var b bytes.Buffer
		reader := newRespReader(bufio.NewReader(upConn), &b, nil)
		if err := reader.ReadReply(); err != nil {
			log.Printf("Error reading subscription response: %v", err)
			conn.Close()
			return
		}

		conn.WriteRaw(b.Bytes())
	}
}
