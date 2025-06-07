package shipnav

import "github.com/moosethebrown/ship-net-bridge/core"

const (
	rqTypeQuery = "query"
	rqTypeCmd   = "cmd"
)

const (
	cmdQuery            = "query"
	cmdNavStart         = "nav_start"
	cmdNavStop          = "nav_stop"
	cmdNetLoss          = "net_loss"
	cmdSetWaypoints     = "set_waypoints"
	cmdAddWaypoint      = "add_waypoint"
	cmdClearWaypoints   = "clear_waypoints"
	cmdSetHomeWaypoint  = "set_home_waypoint"
	cmdStartCalibration = "start_calibration"
	cmdStopCalibration  = "stop_calibration"
)

type Request struct {
	Type      string           `json:"type"`
	Cmd       string           `json:"cmd"`
	Waypoints []*core.Waypoint `json:"waypoints"`
}
