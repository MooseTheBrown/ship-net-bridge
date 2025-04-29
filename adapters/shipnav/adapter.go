package shipnav

import (
	"encoding/json"
	"net"

	"github.com/moosethebrown/ship-net-bridge/core"
	"github.com/rs/zerolog"
)

type cmd struct {
	cmd       string
	waypoints []*core.Waypoint
}

type Adapter struct {
	socketName string
	stopChan   chan bool
	cmdChan    chan *cmd
	respBuf    []byte
	theCore    *core.Core
	logger     *zerolog.Logger
}

func NewAdapter(sockeName string, theCore *core.Core, queueSize int, logger *zerolog.Logger) *Adapter {
	return &Adapter{
		socketName: sockeName,
		stopChan:   make(chan bool, 1),
		cmdChan:    make(chan *cmd, queueSize),
		respBuf:    make([]byte, 4096),
		theCore:    theCore,
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
		case c := <-a.cmdChan:
			a.sendMessage(conn, c)
		case <-a.stopChan:
			break main_loop
		}
	}
}

func (a *Adapter) Stop() {
	a.stopChan <- true
}

func (a *Adapter) Query() {
	a.cmdChan <- &cmd{
		cmd: cmdQuery,
	}
}

func (a *Adapter) NavStart() {
	a.cmdChan <- &cmd{
		cmd: cmdNavStart,
	}
}

func (a *Adapter) NavStop() {
	a.cmdChan <- &cmd{
		cmd: cmdNavStop,
	}
}

func (a *Adapter) NetLoss() {
	a.cmdChan <- &cmd{
		cmd: cmdNetLoss,
	}
}

func (a *Adapter) SetWaypoints(waypoints []*core.Waypoint) {
	a.cmdChan <- &cmd{
		cmd:       cmdNetLoss,
		waypoints: waypoints,
	}
}

func (a *Adapter) AddWaypoint(waypoint *core.Waypoint) {
	c := &cmd{
		cmd:       cmdAddWaypoint,
		waypoints: make([]*core.Waypoint, 1),
	}
	c.waypoints[0] = waypoint
	a.cmdChan <- c
}

func (a *Adapter) ClearWaypoints() {
	a.cmdChan <- &cmd{
		cmd: cmdClearWaypoints,
	}
}

func (a *Adapter) SetHomeWaypoint(waypoint *core.Waypoint) {
	c := &cmd{
		cmd:       cmdSetHomeWaypoint,
		waypoints: make([]*core.Waypoint, 1),
	}
	c.waypoints[0] = waypoint
	a.cmdChan <- c
}

func (a *Adapter) sendMessage(conn net.Conn, command *cmd) {
	rq := &Request{}

	if command.cmd == cmdQuery {
		rq.Type = rqTypeQuery
	} else {
		rq.Type = rqTypeCmd
		rq.Cmd = command.cmd
		rq.Waypoints = command.waypoints
	}

	data, err := json.Marshal(rq)
	if err != nil {
		a.logger.Error().Err(err).Msg("failed to marshal query")
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		a.logger.Error().Err(err).Msg("failed to send query")
		return
	}

	n, err := conn.Read(a.respBuf)
	if err != nil {
		a.logger.Error().Err(err).Msg("failed to receive query response")
		return
	}

	a.theCore.HandleResponse(a.respBuf[:n])
}
