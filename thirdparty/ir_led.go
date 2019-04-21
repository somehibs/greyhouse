package thirdparty

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -lpigpio -lm -lpthread
// #include <stdio.h>
// #include "irslinger.h"
import "C"

type IrBlaster struct {
	pin uint32
}

func NewIrBlaster(pin uint32, irKeys map[string]string) IrBlaster {
	blaster := IrBlaster{pin}
	blaster.SendNec("")
	return blaster
}

const frequency = C.int(38000)
const dutyCycle = C.double(0.5)
const leadingPulseDuration = C.int(9000)
const leadingGapDuration = C.int(4500)
const onePulse = C.int(562)
const zeroPulse = C.int(562)
const oneGap = C.int(1688)
const zeroGap = C.int(562)
const sendTrailingPulse = C.int(1)

func (ir *IrBlaster) SendNec(code string) {
	C.irSling(C.uint(ir.pin), frequency, dutyCycle, leadingPulseDuration, leadingGapDuration, onePulse, zeroPulse, oneGap, zeroGap, sendTrailingPulse, C.CString(code))
}
