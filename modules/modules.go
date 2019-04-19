package modules

import (
	api "git.circuitco.de/self/greyhouse/api"
)

type ClientHost struct {
	Node *api.PrimaryNodeClient
	Presence *api.PresenceClient
	Person *api.PersonClient
	Rules *api.RulesClient
}

type GreyhouseClientModule interface {
	// set some config values. keys must be ints exposed by the module to avoid string comparisons or stupid shit like that
	//SetString(int, string) error
	//SetInt(int, int) error
	// update may not have been called but all config should be finished before init is called
	Init() error
	// Force some new clients on this module, due to connection state change or otherwise
	// can be nil!
	Update(*ClientHost)
	// Called once to find out if this module needs per-second processing
	CanTick() bool
	// Called once per second if canTick returns true
	// otherwise, called once every 10 ticks to ensure networking still OK
	Tick() error
	// Shutdown is called when the application is closing for some reason
	Shutdown()
}
