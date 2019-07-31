package modules

import (
	"log"
	"errors"

	api "git.circuitco.de/self/greyhouse/api"
)

var chost *ClientHost
type V4lStreamer struct {
	lastErr error
}

func NewV4lStreamer(pin int16) V4lStreamer {
	return V4lStreamer{nil}
}

func (watch *V4lStreamer) Init() error {
	// Connect to the V4L device.
	devices := v4l.FindDevices()
	log.Printf("Found 
	return nil
}

func (watch *V4lStreamer) Shutdown() {
	log.Print("streamer shutting down")
}

func (watch *V4lStreamer) writeUpdate(img []byte) {
	if chost == nil {
		log.Print("Cannot report image frame to empty chost")
		return
	}
	//ctx := (*chost).GetContext()
	//update := api.PresenceUpdate {
	//	SensorId: "idfk",
	//	Type: api.PresenceType_Motion,
	//	Distance: 0,
	//	Accuracy: 0,
	//	PeopleDetected: int32(peopleDetected),
	//}
	//_, watch.lastErr = (*chost.Presence).Update(ctx, &update)
}

func (watch *V4lStreamer) Update(ch *ClientHost) {
	chost = ch
	//watch.writeUpdate(watch.pin.Read())
}

func (watch *V4lStreamer) clearError() {
	watch.lastErr = nil
}

func (watch *V4lStreamer) CanTick() bool { return false }
func (watch *V4lStreamer) Tick() error {
	defer watch.clearError()
	return watch.lastErr
}
