package shipcontrol

import (
	"net"

	"github.com/moosethebrown/ship-net-bridge/core"
	"github.com/rs/zerolog"
)

type Adapter struct {
	socketName string
	theCore    *core.Core
	rqChan     chan []byte
	stopChan   chan bool
	respBuf    []byte
	logger     *zerolog.Logger
}

func NewAdapter(socketName string, theCore *core.Core, queueSize int, logger *zerolog.Logger) *Adapter {
	return &Adapter{
		socketName: socketName,
		theCore:    theCore,
		rqChan:     make(chan []byte, queueSize),
		stopChan:   make(chan bool, 1),
		respBuf:    make([]byte, 4096),
		logger:     logger,
	}
}

func (a *Adapter) Run() {
	conn, err := net.Dial("unix", a.socketName)
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to connect to socket")
		return
	}
	defer conn.Close()

main_loop:
	for {
		select {
		case msg := <-a.rqChan:
			a.send(conn, msg)
		case <-a.stopChan:
			break main_loop
		}
	}
}

func (a *Adapter) Stop() {
	a.stopChan <- true
}

func (a *Adapter) SendRequest(msg []byte) {
	a.rqChan <- msg
}

func (a *Adapter) send(conn net.Conn, msg []byte) {
	_, err := conn.Write(msg)
	if err != nil {
		a.logger.Error().Err(err).Msg("failed to send message to ship-control")
		return
	}

	n, err := conn.Read(a.respBuf)
	if err != nil {
		a.logger.Error().Err(err).Msg("failed to read response from ship-control")
		return
	}

	a.theCore.HandleResponse(a.respBuf[:n])
}
