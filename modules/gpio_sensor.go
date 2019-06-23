package modules

import (
	"log"
	"errors"

	"github.com/warthog618/gpio"

	api "git.circuitco.de/self/greyhouse/api"
)

var chost *ClientHost
// This module lets you listen to a GPIO pin.
// Based on type, we'll report to the right API.
// For now, the only thing that raises and lowers GPIO is the PIR sensors.
type GpioWatcher struct {
	pinId int16
	pin *gpio.Pin
	reportHigh bool
	reportLow bool
	lastErr error
}

func NewGpioWatcher(pin int16) GpioWatcher {
	return GpioWatcher{pin, nil, true, true, nil}
}

func (watch *GpioWatcher) Init() error {
	err := gpio.Open()
	if err != nil {
		return err
	}
	watch.pin = gpio.NewPin(uint8(watch.pinId))
	if watch.pin == nil {
		return errors.New("cannot open pin")
	}
	err = watch.pin.Watch(gpio.EdgeBoth, watch.handle)
	return err
}

func (watch *GpioWatcher) Shutdown() {
	log.Printf("gpio watcher shutting down for pin %d", watch.pinId)
	watch.pin.Unwatch()
}

func (watch *GpioWatcher) handle(pin *gpio.Pin) {
	pinState := pin.Read()
	peopleDetected := 0
	if pinState {
		peopleDetected += 1
	}
	log.Printf("Pin %s is %+v", watch.pinId, pinState)
	// Tell someone the pin changed
	watch.writeUpdate(pin.Read())
}

func (watch *GpioWatcher) writeUpdate(pinState gpio.Level) {
	if chost == nil {
		log.Print("Cannot report pin change to empty chost")
		return
	}
	ctx := (*chost).GetContext()
	peopleDetected := 0
	if pinState {
		peopleDetected = 1
	}
	update := api.PresenceUpdate {
		SensorId: "idfk",
		Type: api.PresenceType_Motion,
		Distance: 0,
		Accuracy: 0,
		PeopleDetected: int32(peopleDetected),
	}
	(*chost.Presence).Update(ctx, &update)
}

func (watch *GpioWatcher) Update(ch *ClientHost) {
	chost = ch
	watch.writeUpdate(watch.pin.Read())
}

func (watch *GpioWatcher) CanTick() bool { return false }
func (watch *GpioWatcher) Tick() error {
	return nil
}
