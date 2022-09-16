package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

const (
	MQTT_INTERRUPT_CMD = "!$interrupt$!"
)

type MqttHandler struct {
	in                chan string
	out               chan string
	broker            string
	connTimeout       time.Duration
	username          string
	password          string
	shipid            string
	announceTopic     string
	announceTimeout   time.Duration
	disconnectTimeout time.Duration
	rqTopic           string
	respTopic         string
	certCheck         bool
}

func NewMqttHandler(in chan string, out chan string, broker string, cTimeout time.Duration,
	username string, password string, shipid string, announce string,
	aTimeout time.Duration, dTimeout time.Duration, certCheck bool) *MqttHandler {

	return &MqttHandler{
		in:                in,
		out:               out,
		broker:            broker,
		connTimeout:       cTimeout,
		username:          username,
		password:          password,
		shipid:            shipid,
		announceTopic:     announce,
		announceTimeout:   aTimeout,
		disconnectTimeout: dTimeout,
		rqTopic:           fmt.Sprintf("ship/%s/request", shipid),
		respTopic:         fmt.Sprintf("ship/%s/response", shipid),
		certCheck:         certCheck,
	}
}

// handler's main loop
func (handler *MqttHandler) Run() {
	// setup MQTT client and connect to broker
	opts := mqtt.NewClientOptions().AddBroker(handler.broker).SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetCredentialsProvider(func() (username string, password string) {
		return handler.username, handler.password
	})
	opts.SetClientID(handler.shipid)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !handler.certCheck,
	}
	opts.SetTLSConfig(tlsConfig)

	client := mqtt.NewClient(opts)
	token := client.Connect()

	if token.WaitTimeout(handler.connTimeout) == false {
		panic("MqttHandler failed to connect to broker")
	}
	if err := token.Error(); err != nil {
		panic(err)
	}

	defer client.Disconnect(uint(handler.disconnectTimeout.Nanoseconds() / 1000000))

	// schedule announce message to be published periodically
	ticker := time.NewTicker(handler.announceTimeout)
	defer ticker.Stop()

	// subscribe to request topic
	client.Subscribe(handler.rqTopic, 2, func(cl mqtt.Client, msg mqtt.Message) {
		rq := bytes.NewBuffer(msg.Payload()).String()
		handler.out <- rq
	})

	// send responses using response topic
mqtt_loop:
	for {
		select {
		case resp := <-handler.in:
			// send response to the broker
			if resp == MQTT_INTERRUPT_CMD {
				break mqtt_loop
			}
			client.Publish(handler.respTopic, 2, false, resp)
		case <-ticker.C:
			// time to announce
			client.Publish(handler.announceTopic, 2, false, handler.shipid)
		}
	}
}
