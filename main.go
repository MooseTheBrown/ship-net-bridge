package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"
)

type NetHandler interface {
	Run()
}

func main() {
	// configure
	var configfile string
	flag.StringVar(&configfile, "c", "/etc/ship-net-bridge.conf", "path to configuration file")
	flag.Parse()

	config, err := ReadConfig(configfile)
	if err != nil {
		log.Printf("Error reading config: %s\n", err)
		return
	}

	// launch handlers
	unixin := make(chan string, 10)
	unixout := make(chan string, 10)
	unixHandler := NewUnixHandler(config.UnixSocket, unixin, unixout)
	go runHandler(unixHandler)

	mqttin := make(chan string, 10)
	mqttout := make(chan string, 10)
	mqttHandler := NewMqttHandler(mqttin, mqttout, config.MqttBroker,
		time.Duration(config.ConnTimeout)*time.Millisecond, config.BrokerUsername, config.BrokerPassword,
		config.ShipId, config.AnnounceTopic, time.Duration(config.AnnounceTimeout)*time.Millisecond,
		time.Duration(config.DisconnectTimeout)*time.Millisecond, config.BrokerCertCheck)
	go runHandler(mqttHandler)

	// set graceful exit procedure
	defer func() {
		log.Println("main shutting down")
		e := recover()
		if e != nil {
			log.Printf("Panic caught in main: %s, exiting\n", e)
		}

		// send interrupt commands to the handlers
		unixin <- UNIX_INTERRUPT_CMD
		mqttin <- MQTT_INTERRUPT_CMD
	}()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

main_loop:
	for {
		select {
		case rq := <-mqttout:
			unixin <- rq
		case resp := <-unixout:
			mqttin <- resp
		case <-sigch:
			// SIGINT caught, stop everything and quit
			log.Println("SIGINT")
			unixin <- UNIX_INTERRUPT_CMD
			mqttin <- MQTT_INTERRUPT_CMD
			time.Sleep(100 * time.Millisecond)
			break main_loop
		}
	}
}

// run a NetHandler in a separate goroutine restarting it if it panics
func runHandler(handler NetHandler) {
	restart := make(chan bool, 1)

	for {
		go func() {
			defer func() {
				err := recover()
				if err != nil {
					log.Printf("NetHandler panicked: %s\n", err)
				}
				restart <- true
			}()

			handler.Run()
			// if Run() exited normally, there's no need to restart
			restart <- false
		}()

		r := <-restart
		if r == true {
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
}
