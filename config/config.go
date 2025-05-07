package config

import (
	"encoding/json"
	"io"
	"os"
)

type MqttConfig struct {
	Broker            string `json:"broker"`
	ConnTimeout       int    `json:"connTimeout"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	ShipId            string `json:"shipId"`
	AnnounceTopic     string `json:"announceTopic"`
	AnnounceTimeout   int    `json:"announceTimeout"`
	DisconnectTimeout int    `json:"disconnectTimeout"`
	CertCheck         bool   `json:"certCheck"`
}

type ShipControlConfig struct {
	SocketName string `json:"socketName"`
	QueueSize  int    `json:"queueSize"`
}

type ShipNavConfig struct {
	SocketName string `json:"socketName"`
	QueueSize  int    `json:"queueSize"`
}

// JSON-based bridge configuration
type Config struct {
	Mqtt             *MqttConfig        `json:"mqtt"`
	ShipControl      *ShipControlConfig `json:"shipControl"`
	ShipNav          *ShipNavConfig     `json:"shipNav"`
	AnnounceInterval int                `json:"announceInterval"`
	LogLevel         string             `json:"logLevel"`
}

func NewConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
