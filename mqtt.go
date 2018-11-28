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
	shipid            string
	announceTopic     string
	announceTimeout   time.Duration
	disconnectTimeout time.Duration
	rqTopic           string
	respTopic         string
	certCheck         bool
}

func NewMqttHandler(in chan string, out chan string, broker string, cTimeout time.Duration,
	shipid string, announce string, aTimeout time.Duration, dTimeout time.Duration,
	certCheck bool) *MqttHandler {

	return &MqttHandler{
		in:                in,
		out:               out,
		broker:            broker,
		connTimeout:       cTimeout,
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

	// signal that the ship is able to communicate by sending shipid through announce topic
	// and wait until the message is delivered
	// NOTE: announce message should be retained by broker so that any control software starting
	// AFTER this announce will get it
	pubToken := client.Publish(handler.announceTopic, 2, true, handler.shipid)
	if pubToken.WaitTimeout(handler.announceTimeout) == false {
		panic("MqttHandler failed to announce")
	}
	if err := pubToken.Error(); err != nil {
		panic(err)
	}

	// subscribe to request topic
	client.Subscribe(handler.rqTopic, 2, func(cl mqtt.Client, msg mqtt.Message) {
		rq := bytes.NewBuffer(msg.Payload()).String()
		handler.out <- rq
	})

	// send responses using response topic
	for {
		resp := <-handler.in
		if resp == MQTT_INTERRUPT_CMD {
			break
		}
		client.Publish(handler.respTopic, 2, false, resp)
	}
}
