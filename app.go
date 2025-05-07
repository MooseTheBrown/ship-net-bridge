package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/moosethebrown/ship-net-bridge/adapters/mqtt"
	"github.com/moosethebrown/ship-net-bridge/adapters/shipcontrol"
	"github.com/moosethebrown/ship-net-bridge/adapters/shipnav"
	"github.com/moosethebrown/ship-net-bridge/config"
	"github.com/moosethebrown/ship-net-bridge/core"
	"github.com/rs/zerolog"
)

type App struct {
	cfg                *config.Config
	logger             *zerolog.Logger
	mqttAdapter        *mqtt.Adapter
	shipControlAdapter *shipcontrol.Adapter
	shipNavAdapter     *shipnav.Adapter
	theCore            *core.Core
	wg                 sync.WaitGroup
}

func NewApp(cfg *config.Config) *App {
	app := &App{
		cfg: cfg,
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		fmt.Printf("Invalid logLevel: %s, error: %s", cfg.LogLevel, err.Error())
		logLevel = zerolog.InfoLevel
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(logLevel)
	app.logger = &logger

	app.init()

	return app
}

func (app *App) Start() {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		app.theCore.Run()
	}()

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		app.shipControlAdapter.Run()
	}()

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		app.shipNavAdapter.Run()
	}()

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		err := app.mqttAdapter.Run()
		if err != nil {
			panic("mqtt adapter exited unexpectedly")
		}
	}()
}

func (app *App) Stop() {
	app.mqttAdapter.Stop()
	app.shipNavAdapter.Stop()
	app.shipControlAdapter.Stop()
	app.theCore.Stop()
	app.wg.Wait()
}

func (app *App) init() {
	coreLogger := app.logger.With().Str("component", "core").Logger()
	app.theCore = core.NewCore(nil, nil, nil,
		app.cfg.AnnounceInterval, &coreLogger)

	mqttLogger := app.logger.With().Str("component", "mqtt").Logger()
	app.mqttAdapter = mqtt.NewAdapter(app.cfg.Mqtt.Broker,
		time.Duration(app.cfg.Mqtt.ConnTimeout)*time.Millisecond,
		app.cfg.Mqtt.Username,
		app.cfg.Mqtt.Password,
		app.cfg.Mqtt.ShipId,
		app.cfg.Mqtt.AnnounceTopic,
		time.Duration(app.cfg.Mqtt.AnnounceTimeout)*time.Millisecond,
		time.Duration(app.cfg.Mqtt.DisconnectTimeout)*time.Millisecond,
		app.cfg.Mqtt.CertCheck,
		app.theCore,
		&mqttLogger)
	app.theCore.SetMqttHandler(app.mqttAdapter)

	shipControlLogger := app.logger.With().Str("component", "ship-control").Logger()
	app.shipControlAdapter = shipcontrol.NewAdapter(app.cfg.ShipControl.SocketName,
		app.theCore,
		app.cfg.ShipControl.QueueSize,
		&shipControlLogger)
	app.theCore.SetShipControl(app.shipControlAdapter)

	shipNavLogger := app.logger.With().Str("component", "ship-nav").Logger()
	app.shipNavAdapter = shipnav.NewAdapter(app.cfg.ShipNav.SocketName,
		app.theCore,
		app.cfg.ShipNav.QueueSize,
		&shipNavLogger)
	app.theCore.SetShipNav(app.shipNavAdapter)
}
