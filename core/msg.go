package core

const (
	RequestTypeCmd   = "cmd"
	RequestTypeQuery = "query"
)

const (
	CmdSpeedUp         = "speed_up"
	CmdSpeedDown       = "speed_down"
	CmdTurnLeft        = "turn_left"
	CmdTurnRight       = "turn_right"
	CmdSetSpeed        = "set_speed"
	CmdSetSteering     = "set_steering"
	CmdSetWaypoints    = "set_waypoints"
	CmdAddWaypoint     = "add_waypoint"
	CmdClearWaypoints  = "clear_waypoints"
	CmdSetHomeWaypoint = "set_home_waypoint"
	CmdNavStart        = "nav_start"
	CmdNetLoss         = "net_loss"
)

type Waypoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Request struct {
	Type      string `json:"type"`
	Cmd       string `json:"cmd"`
	Data      string `json:"data"`
	rawData   []byte
	waypoints []*Waypoint
}
