package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/moosethebrown/ship-net-bridge/config"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "/etc/ship-net-bridge.conf", "path to configuration file")
	flag.Parse()

	cfg, err := config.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error reading config: %s\n", err)
		return
	}

	app := NewApp(cfg)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

	app.Start()

	<-sigch
	app.Stop()
}
