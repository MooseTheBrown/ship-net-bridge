package core

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

type mockShipControl struct {
}

func (m *mockShipControl) SendRequest([]byte) {
}

type mockShipNav struct {
}

func (m *mockShipNav) Query() {
}

func (m *mockShipNav) NavStart() {
}

func (m *mockShipNav) NavStop() {
}

func (m *mockShipNav) NetLoss() {
}

func (m *mockShipNav) SetWaypoints([]*Waypoint) {
}

func (m *mockShipNav) AddWaypoint(*Waypoint) {
}

func (m *mockShipNav) ClearWaypoints() {
}

func (m *mockShipNav) SetHomeWaypoint(*Waypoint) {
}

func (m *mockShipNav) StartCalibration() {
}

func (m *mockShipNav) StopCalibration() {
}

type mockMqttHandler struct {
}

func (m *mockMqttHandler) SendResponse([]byte) {
}

func (m *mockMqttHandler) Announce() {
}

func setup() *Core {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	core := NewCore(&mockShipControl{}, &mockShipNav{},
		&mockMqttHandler{}, 3000, &logger)

	return core
}

func TestWaypointsParsing(t *testing.T) {
	core := setup()

	rq := &Request{
		Type: RequestTypeCmd,
		Cmd:  CmdSetWaypoints,
		Data: "56.348284,43.959410;56.359226,43.907618",
	}

	core.parseWaypoints(rq)

	if len(rq.waypoints) != 2 {
		t.Fatalf("Expected 2 waypoints, got %d", len(rq.waypoints))
	}

	if rq.waypoints[0].Latitude != 56.348284 {
		t.Errorf("Expected waypoint 1 latitude to be 56.348284, got %f",
			rq.waypoints[0].Latitude)
	}
	if rq.waypoints[0].Longitude != 43.959410 {
		t.Errorf("Expected waypoint 1 longitude to be 43.959410, got %f",
			rq.waypoints[0].Longitude)
	}
	if rq.waypoints[1].Latitude != 56.359226 {
		t.Errorf("Expected waypoint 2 latitude to be 56.359226, got %f",
			rq.waypoints[1].Latitude)
	}
	if rq.waypoints[1].Longitude != 43.907618 {
		t.Errorf("Expected waypoint 2 longitude to be 43.907618, got %f",
			rq.waypoints[1].Longitude)
	}
}
