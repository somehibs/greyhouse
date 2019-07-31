package modules

import (
	"log"
	"errors"
	"time"

	api "git.circuitco.de/self/greyhouse/api"

	"github.com/korandiz/v4l"

	"git.circuitco.de/self/greyhouse/recognise"
)

type V4lStreamer struct {
	device *v4l.Device
	lastErr error
	lastUpload *time.Time
	recogniser *recognise.Recogniser
}

func NewV4lStreamer() V4lStreamer {
	return V4lStreamer{nil, nil, nil, nil}
}

func (s *V4lStreamer) Init() error {
	// load up tensorflow, fuck it, why not
	recogniser := recognise.NewRecogniser("model")
	s.recogniser = &recogniser
	// Check for V4L devices
	devices := v4l.FindDevices()
	if len(devices) > 0 {
		// Connect to the V4L device.
		log.Printf("Using first device out of %d", len(devices))
		deviceInfo := devices[0]
		device, err := v4l.Open(deviceInfo.Path)
		s.device = device
		if err != nil {
			log.Printf("Err opening device: %+v", err)
			return err
		}
		// Check device config
		_, err = device.GetConfig()
		if err != nil {
			log.Printf("Err fetching device config: %+v", err)
			return err
		}
		//log.Printf("gcf: %+v", gcf)
		if false {
			// TODO: make GetConfig empty work
			cfg, err := device.ListConfigs()
			if err != nil {
				log.Printf("Err listing device config: %+v", err)
				return err
			}
			log.Printf("Config available: %+v", cfg)
		} else {
			// v4l is already running, what's the config?
			bufferInfo, err := device.BufferInfo()
			if err != nil {
				return err
			}
			log.Printf("Buffer info: %+v", bufferInfo)
			// enable our connection to the device
			err = device.TurnOn()
			if err != nil {
				return err
			}
			return err
		}
	} else {
		log.Printf("Could not find any devices for streamer.")
		return errors.New("v4l_no_devices")
	}
	return nil
}

func (s *V4lStreamer) CaptureFrame() (*v4l.Buffer, time.Time, error) {
	start := time.Now()
	buffer, err := s.device.Capture()
	end := time.Now()
	log.Printf("Capture time %+v", end.Sub(start))
	log.Printf("Buf %+v", buffer)
	log.Printf("Err %s", err)
	return buffer, start, err
}

func (s *V4lStreamer) Shutdown() {
	log.Print("streamer shutting down")
	if s.device != nil {
		s.device.Close()
	}
}

func (watch *V4lStreamer) writeUpdate(t time.Time, img *v4l.Buffer) {
	if chost == nil {
		log.Print("Cannot report image frame to empty chost")
		return
	}
	ctx := (*chost).GetContext()
	update := api.ImageUpdate {
		Time: t.Unix(),
		Image: img.Source(),
	}
	//	Distance: 0,
	//	Accuracy: 0,
	//	PeopleDetected: int32(peopleDetected),
	//}
	start := time.Now()
	_, watch.lastErr = (*chost.Presence).Image(ctx, &update)
	end := time.Now()
	log.Printf("Uploading took %s", end.Sub(start))
}

func (s *V4lStreamer) Update() {
	frame, t, err := s.CaptureFrame()
	if err != nil {
		log.Panicf("Cannot capture frame %+v", err)
	}
	s.writeUpdate(t, frame)
}

func (watch *V4lStreamer) clearError() {
	watch.lastErr = nil
}

func (watch *V4lStreamer) CanTick() bool { return true }
func (watch *V4lStreamer) Tick() error {
	defer watch.clearError()
	watch.Update()
	return watch.lastErr
}
