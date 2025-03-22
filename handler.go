package main

import (
	"bufio"
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

	// Modify key (if applicable)
	if modCommands[command] && len(cmd.Args) > 1 {
		cmd.Args[1] = append([]byte(context.username+":"), cmd.Args[1]...)
	}

	// Construct RESP command
	request := buildRESPCommand(cmd.Args)

	// Send to Redis
	_, err := upConn.Write(request)
	if err != nil {
		log.Printf("Failed to send command: %v", err)
		conn.Close()
		return
	}

	// Read response from Redis
	response, err := bufio.NewReader(upConn).ReadBytes('\n')
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		conn.Close()
		return
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
	var sb strings.Builder
	sb.WriteString("*")
	sb.WriteString(strings.TrimSpace(string([]byte{byte(len(args) + '0')})))
	sb.WriteString("\r\n")
	for _, arg := range args {
		sb.WriteString("$")
		sb.WriteString(strings.TrimSpace(string([]byte{byte(len(arg) + '0')})))
		sb.WriteString("\r\n")
		sb.Write(arg)
		sb.WriteString("\r\n")
	}
	return []byte(sb.String())
}
