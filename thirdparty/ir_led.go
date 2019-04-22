package thirdparty

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -lpigpio -lm -lpthread
// #include <stdio.h>
// #include "irslinger.h"
import "C"

import "fmt"

type IrBlaster struct {
	pin uint32
	keys map[string]string
}

func NewIrBlaster(pin uint32, irKeys map[string]string) IrBlaster {
	blaster := IrBlaster{pin, irKeys}
	blaster.SendNec("")
	return blaster
}

func (ib *IrBlaster) SendKey(key string) {
	fmt.Printf("sending key %s - %s", key, ib.keys[key])
	ib.SendNec(ib.keys[key])
}

const frequency = C.int(38000)
const dutyCycle = C.double(0.5)
const leadingPulseDuration = C.int(9000)
const leadingGapDuration = C.int(4500)
const onePulse = C.int(550)
const zeroPulse = C.int(550)
const oneGap = C.int(1700)
const zeroGap = C.int(500)
const sendTrailingPulse = C.int(0)

func (ir *IrBlaster) SendNec(code string) {
	C.irSling(C.uint(ir.pin), frequency, dutyCycle, leadingPulseDuration, leadingGapDuration, onePulse, zeroPulse, oneGap, zeroGap, sendTrailingPulse, C.CString(code))
}
