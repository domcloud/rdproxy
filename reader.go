package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type transformerFn func(code byte, line []byte) []byte

func newRespReader(b *bufio.Reader, w *bytes.Buffer, fn transformerFn) RespReader {
	return RespReader{
		br: b,
		bw: w,
		fn: fn,
	}
}

type RespReader struct {
	br *bufio.Reader
	bw *bytes.Buffer
	fn transformerFn
}

type readerError string

func (pe readerError) Error() string {
	return fmt.Sprintf("parse erro: %s", string(pe))
}

func (c *RespReader) readReply() error {
	line, err := c.readLine()
	if err != nil {
		return err
	}
	if len(line) < 2 {
		return readerError("short response line")
	}
	switch line[0] {
	case '+', '-', ':':
		c.bw.Write(line)
		return nil
	case '$':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return err
		}
		p := make([]byte, n+2)
		_, err = io.ReadFull(c.br, p)
		if err != nil {
			return err
		}
		if c.fn != nil {
			p = c.fn('$', p)
		}
		c.bw.WriteString("$" + strconv.Itoa(len(p)-2) + "\r\n")
		c.bw.Write(p)
		return nil
	case '*':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return err
		}
		c.bw.Write(line)
		for range n {
			err = c.readReply()
			if err != nil {
				return err
			}
		}
		return nil
	}
	return readerError("unexpected response line")
}

func (c *RespReader) readLine() ([]byte, error) {
	// To avoid allocations, attempt to read the line using ReadSlice. This
	// call typically succeeds. The known case where the call fails is when
	// reading the output from the MONITOR command.
	p, err := c.br.ReadSlice('\n')
	if err == bufio.ErrBufferFull {
		// The line does not fit in the bufio.Reader's buffer. Fall back to
		// allocating a buffer for the line.
		buf := append([]byte{}, p...)
		for err == bufio.ErrBufferFull {
			p, err = c.br.ReadSlice('\n')
			buf = append(buf, p...)
		}
		p = buf
	}
	if err != nil {
		return nil, err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return nil, readerError("bad response line terminator")
	}
	return p, nil
}

// parseLen parses bulk string and array lengths.
func parseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, readerError("malformed length")
	}

	if p[0] == '-' && len(p) == 4 && p[1] == '1' {
		// handle $-1 and $-1 null replies.
		return -1, nil
	}

	var n int
	for _, b := range p[:len(p)-2] {
		n *= 10
		if b < '0' || b > '9' {
			return -1, readerError("illegal bytes in length")
		}
		n += int(b - '0')
	}

	return n, nil
}
