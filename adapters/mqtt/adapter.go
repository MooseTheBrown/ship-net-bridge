package mqtt

import (
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/moosethebrown/ship-net-bridge/core"
	"github.com/rs/zerolog"
)

type Adapter struct {
	broker            string
	connTimeout       time.Duration
	username          string
	passwd            string
	shipId            string
	announceTopic     string
	announceTimeout   time.Duration
	disconnectTimeout time.Duration
	rqTopic           string
	respTopic         string
	certCheck         bool
	client            mqtt.Client
	core              *core.Core
	stopChan          chan bool
	announceChan      chan bool
	responseChan      chan []byte
	logger            *zerolog.Logger
}

func NewAdapter(broker string, connTimeout time.Duration, username string,
	passwd string, shipId string, announceTopic string,
	announceTimeout time.Duration,
	disconnectTimeout time.Duration,
	certCheck bool, core *core.Core,
	logger *zerolog.Logger) *Adapter {

	return &Adapter{
		broker:            broker,
		connTimeout:       connTimeout,
		username:          username,
		passwd:            passwd,
		shipId:            shipId,
		announceTopic:     announceTopic,
		announceTimeout:   announceTimeout,
		disconnectTimeout: disconnectTimeout,
		rqTopic:           fmt.Sprintf("ship/%s/request", shipId),
		respTopic:         fmt.Sprintf("ship/%s/response", shipId),
		certCheck:         certCheck,
		core:              core,
		stopChan:          make(chan bool, 1),
		announceChan:      make(chan bool, 1),
		responseChan:      make(chan []byte, 1000),
		logger:            logger,
	}
}

func (a *Adapter) Run() error {
	a.logger.Info().Msg("starting")
	defer a.logger.Info().Msg("stopping")

	err := a.connect()
	if err != nil {
		a.logger.Error().Err(err).Msg("failed to connect to MQTT broker")
		return err
	}

	defer a.client.Disconnect(uint(a.disconnectTimeout.Milliseconds()))

main_loop:
	for {
		select {
		case <-a.stopChan:
			break main_loop
		case <-a.announceChan:
			a.logger.Debug().Msg("announce")

			token := a.client.Publish(a.announceTopic, 2, false, a.shipId)
			// TODO: maybe report net loss only after N consecutive failed announce attempts?
			if token.WaitTimeout(a.announceTimeout) == false {
				a.logger.Error().Msg("timeout expired while publishing announce message")
				a.core.NetLoss()
			} else if err := token.Error(); err != nil {
				a.logger.Error().Err(err).Msg("error publishing announce message")
				a.core.NetLoss()
			}
		case resp := <-a.responseChan:
			a.client.Publish(a.respTopic, 2, false, resp)
		}
	}

	return nil
}

func (a *Adapter) Stop() {
	a.stopChan <- true
}

func (a *Adapter) SendResponse(resp []byte) {
	a.responseChan <- resp
}

func (a *Adapter) Announce() {
	a.announceChan <- true
}

func (a *Adapter) connect() error {
	opts := mqtt.NewClientOptions().AddBroker(a.broker).SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetCredentialsProvider(func() (username string, password string) {
		return a.username, a.passwd
	})
	opts.SetClientID(a.shipId)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !a.certCheck,
	}
	opts.SetTLSConfig(tlsConfig)
	opts.SetOnConnectHandler(func(cl mqtt.Client) {
		// subscribe to request topic
		cl.Subscribe(a.rqTopic, 2, func(cl mqtt.Client, msg mqtt.Message) {
			a.logger.Debug().Msgf("received request: %s", string(msg.Payload()))
			a.core.HandleRequest(msg.Payload())
		})
	})

	a.client = mqtt.NewClient(opts)
	token := a.client.Connect()

	if token.WaitTimeout(a.connTimeout) == false {
		return errors.New("failed to connect to broker")
	}

	err := token.Error()
	return err
}
