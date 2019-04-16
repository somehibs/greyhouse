package modules

import (
	"log"
	"errors"

	"github.com/warthog618/gpio"
)

var chost ClientHost
// This module lets you listen to a GPIO pin.
// Based on type, we'll report to the right API.
// For now, the only thing that raises and lowers GPIO is the PIR sensors.
type GpioWatcher struct {
	pinId int16
	pin *gpio.Pin
	reportHigh bool
	reportLow bool
}

func NewGpioWatcher(pin int16) GpioWatcher {
	return GpioWatcher{pin, nil, true, true}
}

func (watch GpioWatcher) Init() error {
	err := gpio.Open()
	if err != nil {
		return err
	}
	watch.pin = gpio.NewPin(uint8(watch.pinId))
	if watch.pin == nil {
		return errors.New("cannot open pin")
	}
	watch.pin.Watch(gpio.EdgeBoth, watch.handle)
	return err
}

func (watch GpioWatcher) handle(pin *gpio.Pin) {
	log.Printf("Pin %s is %+v", watch.pinId, pin.Read())
}

func (watch GpioWatcher) handleDown(pin *gpio.Pin) {
	log.Printf("Pin low %s", watch.pinId)
}

func (watch GpioWatcher) Update(ch *ClientHost) {
}

func (watch GpioWatcher) CanTick() bool { return false }
func (watch GpioWatcher) Tick() error { return nil }
