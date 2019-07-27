package modules

import (
	"log"

	"github.com/nicholasjackson/rcswitch"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	//api "git.circuitco.de/self/greyhouse/api"
)

var rfchost *ClientHost
// This module lets you broadcast or receive codes from a 433mhz receiver/sender.
type RfControl struct {
	txPin int16
	rxPin int16

	rTxPin *rcswitch.RCSwitch
	rRxPin *rcswitch.RCSwitch
}

func NewRfControl(txPin int16) RfControl {
	return NewRfControlRx(txPin, 0)
}

func NewRfControlRx(txPin int16, rxPin int16) RfControl {
	log.Print("Ok")
	return RfControl{txPin, rxPin, nil, nil}
}

func (c *RfControl) Init() error {
	_, err := host.Init()
	if err != nil {
		return err
	}
	c.rTxPin = rcswitch.New(gpioreg.ByName(c.txPin))
	if rxPin != 0 {
		c.rRxPin = gpioreg.ByName(c.rxPin)
	}
	return nil
}

func (c *RfControl) Shutdown() {

}

func (c *RfControl) Update(ch *ClientHost) {
	rfchost = ch
}

func (c *RfControl) CanTick() bool { return false }
func (c *RfControl) Tick() error {
	return nil
}
