package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// JSON-based bridge configuration
type Config struct {
	UnixSocket      string
	MqttBroker      string
	BrokerCertCheck bool
	ShipId          string
	AnnounceTopic   string
	// times are in milliseconds
	ConnTimeout       int64
	AnnounceTimeout   int64
	DisconnectTimeout int64
}

func ReadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
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
