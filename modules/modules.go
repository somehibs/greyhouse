package modules

import (
	"log"
	"time"

	"golang.org/x/net/context"
	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/node"
)

type ClientHost struct {
	Key string
	Node *api.PrimaryNodeClient
	Presence *api.PresenceClient
	Person *api.PersonClient
	Rules *api.RulesClient
}

type ModuleConfig struct {
	Name string
	Args map[string]interface{}
}

var chost *ClientHost

func LoadModules(moduleConfig []ModuleConfig) ([]GreyhouseClientModule, error) {
	loaded := make([]GreyhouseClientModule, 0)
	var err error
	for _, config := range moduleConfig {
		var module GreyhouseClientModule
		switch config.Name {
		case "gpio":
			gpio := NewGpioWatcher()
			module = &gpio
		case "video":
			video := NewV4lStreamer()
			module = &video
		default:
			log.Panicf("Module name not recognised: %s", config.Name)
		}
		if module != nil {
			err = module.Init(config)
			if err != nil {
				for _, l := range loaded {
					l.Shutdown()
				}
				return nil, err
			}
			loaded = append(loaded, module)
		}
	}
	return loaded, nil
}

func SetClientHost(_chost *ClientHost) {
	chost = _chost
}

func (c ClientHost) GetContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 9*time.Second)
	return node.AuthContext(ctx, c.Key)
}

type GreyhouseClientModule interface {
	// init takes a moduleconfig and starts up the component
	// init should gracefully return errors and not panic.
	// errors will halt the application safely
	Init(ModuleConfig) error
	// Force some new clients on this module, due to connection state change or otherwise
	// can be nil!
	Update()
	// Called once to find out if this module needs per-second processing
	CanTick() bool
	// Called once per second if canTick returns true
	// otherwise, called once every 10 ticks to ensure networking still OK
	Tick() error
	// Shutdown is called when the application is closing for some reason
	Shutdown()
}

