package main

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"strconv"
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
	reviver := modSingleCommand(command, context.username, cmd.Args)

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
	if err = reader.readReply(); err != nil {
		log.Printf("Failed to read response: %v", err)
		conn.Close()
		return
	}
	response := b.Bytes()

	if command == "AUTH" && len(cmd.Args) == 3 {
		if strings.HasPrefix(string(response), "+OK") {
			context.username = string(cmd.Args[1])
			conn.SetContext(context)
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
	if ok && context.upstreamConn != nil {
		context.upstreamConn.Close()
	}
}

func buildRESPCommand(args [][]byte) []byte {
	var sb bytes.Buffer
	sb.WriteByte('*')
	sb.WriteString(strconv.Itoa(len(args)))
	sb.WriteString("\r\n")
	for _, arg := range args {
		sb.WriteByte('$')
		sb.WriteString(strconv.Itoa(len(arg)))
		sb.WriteString("\r\n")
		sb.Write(arg)
		sb.WriteString("\r\n")
	}
	return sb.Bytes()
}
