package main

import (
	"bufio"
	"net"
)

const (
	UNIX_INTERRUPT_CMD = "!$interrupt$!"
)

// UnixHandler forwards commands from input channel to Unix domain socket
// and forwards responses back through output channel.
type UnixHandler struct {
	in         chan string
	out        chan string
	socketname string
}

func NewUnixHandler(socketname string, in chan string, out chan string) *UnixHandler {
	return &UnixHandler{
		in:         in,
		out:        out,
		socketname: socketname,
	}
}

// Handler main loop.
// May panic.
func (handler *UnixHandler) Run() {
	const READBUF_SIZE = 4096

	conn, err := net.Dial("unix", handler.socketname)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	rqWriter := bufio.NewWriter(conn)

	var rq string
	var resp string
	readbuf := make([]byte, READBUF_SIZE)

	for {
		// forward request to the socket
		rq = <-handler.in
		if rq == UNIX_INTERRUPT_CMD {
			break
		}

		_, err = rqWriter.WriteString(rq)
		if err != nil {
			panic(err)
		}
		rqWriter.Flush()

		// read and forward response
		n, err := conn.Read(readbuf)
		if err != nil {
			panic(err)
		}

		resp = string(readbuf[:n])
		handler.out <- resp
	}
}
