package core

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type ShipControl interface {
	SendRequest([]byte)
}

type ShipNav interface {
	Query()
	NavStart()
	NavStop()
	NetLoss()
	SetWaypoints([]*Waypoint)
	AddWaypoint(*Waypoint)
	ClearWaypoints()
	SetHomeWaypoint(*Waypoint)
}

type MqttHandler interface {
	SendResponse([]byte)
	Announce()
}

type Core struct {
	shipControl      ShipControl
	shipNav          ShipNav
	mqttHandler      MqttHandler
	announceInterval int
	logger           *zerolog.Logger
	rqChan           chan *Request
	respChan         chan []byte
	stopChan         chan bool
	netLossChan      chan bool
	autoNav          bool
}

func NewCore(shipControl ShipControl, shipNav ShipNav,
	mqttHandler MqttHandler, announceInterval int, logger *zerolog.Logger) *Core {
	return &Core{
		shipControl:      shipControl,
		shipNav:          shipNav,
		mqttHandler:      mqttHandler,
		announceInterval: announceInterval,
		logger:           logger,
		rqChan:           make(chan *Request, 1000),
		respChan:         make(chan []byte, 1000),
		stopChan:         make(chan bool, 1),
		netLossChan:      make(chan bool, 1),
	}
}

func (c *Core) SetMqttHandler(handler MqttHandler) {
	c.mqttHandler = handler
}

func (c *Core) SetShipControl(shipControl ShipControl) {
	c.shipControl = shipControl
}

func (c *Core) SetShipNav(shipNav ShipNav) {
	c.shipNav = shipNav
}

func (c *Core) HandleRequest(msg []byte) {
	var rq Request
	rq.waypoints = make([]*Waypoint, 0)

	err := json.Unmarshal(msg, &rq)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to unmarshal request")
		return
	}
	rq.rawData = msg
	c.parseWaypoints(&rq)

	c.rqChan <- &rq

}

func (c *Core) HandleResponse(resp []byte) {
	c.respChan <- resp
}

func (c *Core) Run() {
	ticker := time.NewTicker(time.Duration(time.Duration(c.announceInterval) * time.Millisecond))
	defer ticker.Stop()

core_loop:
	for {
		select {
		case rq := <-c.rqChan:
			if rq.Type == RequestTypeCmd {
				c.handleCommand(rq)
			} else if rq.Type == RequestTypeQuery {
				c.handleQuery()
			} else {
				c.logger.Error().Msgf("unknown request type: %s", rq.Type)
			}
		case resp := <-c.respChan:
			c.mqttHandler.SendResponse(resp)
		case <-ticker.C:
			c.mqttHandler.Announce()
		case <-c.netLossChan:
			c.shipNav.NetLoss()
		case <-c.stopChan:
			break core_loop
		}
	}
}

func (c *Core) Stop() {
	c.stopChan <- true
}

func (c *Core) NetLoss() {
	c.netLossChan <- true
}

func (c *Core) handleCommand(rq *Request) {
	if (rq.Cmd == CmdSpeedUp) || (rq.Cmd == CmdSpeedDown) ||
		(rq.Cmd == CmdTurnLeft) || (rq.Cmd == CmdTurnRight) ||
		(rq.Cmd == CmdSetSpeed) || (rq.Cmd == CmdSetSteering) {
		// control commands go to ship-control directly
		c.shipControl.SendRequest(rq.rawData)
		if c.autoNav {
			c.logger.Info().Msg("received control command, stopping autonav")
			c.shipNav.NavStop()
			c.autoNav = false
		}
	} else if rq.Cmd == CmdSetWaypoints {
		if len(rq.waypoints) == 0 {
			c.logger.Error().Msgf("no waypoints provided for set_waypoints command")
			return
		}
		c.shipNav.SetWaypoints(rq.waypoints)
	} else if rq.Cmd == CmdAddWaypoint {
		if len(rq.waypoints) == 0 {
			c.logger.Error().Msgf("no waypoints provided for add_waypoint command")
			return
		}
		c.shipNav.AddWaypoint(rq.waypoints[0])
	} else if rq.Cmd == CmdClearWaypoints {
		c.shipNav.ClearWaypoints()
	} else if rq.Cmd == CmdSetHomeWaypoint {
		if len(rq.waypoints) == 0 {
			c.logger.Error().Msgf("no waypoints provided for set_home_waypoint command")
			return
		}
		c.shipNav.SetHomeWaypoint(rq.waypoints[0])
	} else if rq.Cmd == CmdNavStart {
		c.shipNav.NavStart()
		c.autoNav = true
	} else {
		c.logger.Error().Msgf("unknown command: %s", rq.Cmd)
	}
}

func (c *Core) handleQuery() {
	c.shipNav.Query()
}

func (c *Core) parseWaypoints(rq *Request) {
	if rq.Type != RequestTypeCmd {
		return
	}
	if (rq.Cmd != CmdSetWaypoints) &&
		(rq.Cmd != CmdAddWaypoint) &&
		(rq.Cmd != CmdSetHomeWaypoint) {
		return
	}

	for _, locstr := range strings.Split(rq.Data, ";") {
		var wp Waypoint
		loc := strings.Split(locstr, ",")
		if len(loc) == 2 {
			var err error
			wp.Latitude, err = strconv.ParseFloat(loc[0], 64)
			if err != nil {
				c.logger.Error().Err(err).Msg("failed to parse waypoint latitude")
			}

			wp.Longitude, err = strconv.ParseFloat(loc[1], 64)
			if err != nil {
				c.logger.Error().Err(err).Msg("failed to parse waypoint latitude")
			}

			rq.waypoints = append(rq.waypoints, &wp)
		}
	}
}
