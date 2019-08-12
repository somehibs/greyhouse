package modules

import (
	"log"
	"errors"
	"time"
	"net/http"

	api "git.circuitco.de/self/greyhouse/api"

	"github.com/korandiz/v4l"
)

func FourCC(cc []byte) uint32 {
	log.Printf("Trying to check byte array: %+v", cc)
	return u32(cc)
}

func u32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

type V4lStreamer struct {
	device *v4l.Device
	lastErr error
	lastUpload *time.Time
	Throttle int32
	UploadsEnabled bool
}

func NewV4lStreamer() V4lStreamer {
	return V4lStreamer{nil, nil, nil, 0, true}
}

func (s *V4lStreamer) restarter() {
	for ;; {
		time.Sleep(1*time.Second)
		s.device.TurnOff()
		s.device.TurnOn(true)
	}
}

func (s *V4lStreamer) Init(config ModuleConfig) error {
	if config.Args["DisableUploads"] != nil && (config.Args["DisableUploads"]).(bool) {
		s.UploadsEnabled = false
	}
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
		gcf, err := device.GetConfig()
		if err != nil {
			log.Printf("Err fetching device config: %+v", err)
			return err
		}
		log.Printf("gcf: %+v", gcf)
		const yuyv_h264 = 875967048
			// TODO: make GetConfig empty work
			cfg, err := device.ListConfigs()
			if err != nil {
				log.Printf("Err listing device config: %+v", err)
				// manually set the device config to what we want
				err := device.SetConfig(v4l.DeviceConfig{Width: 640, Height: 480, Format: FourCC([]byte{'M','J','P','G'}), FPS: v4l.Frac{15, 1}})
				if err != nil {
					log.Printf("Failed to set config: %s", err)
				}
		gcf, err := device.GetConfig()
		log.Printf("gcf: %+v", gcf)
				err = device.TurnOn(false)
				s.listenHttp()
				return err
			} else {
				log.Printf("Config available: %+v", cfg)
			}
	} else {
		log.Printf("Could not find any devices for streamer.")
		return errors.New("v4l_no_devices")
	}
	return nil
}

func (s *V4lStreamer) listenHttp() {
	server := &http.Server {
		Addr: ":80",
		ReadTimeout: 5*time.Second,
		WriteTimeout: 5*time.Second,
	}
	server.SetKeepAlivesEnabled(false)
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		s.device.TurnOn(true)
		b, e := s.device.Capture()
		if e == nil {
			w.Write(b.Source())
		}
		s.device.TurnOff()
	})
	go server.ListenAndServe()
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

func (s *V4lStreamer) writeUpdate(t time.Time, img *v4l.Buffer) {
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
	if chost == nil {
		return
	}
	reply, err := (*chost.Presence).Image(ctx, &update)
	s.lastErr = err
	if err == nil && reply.Throttle != 0 {
		// Throttled for n seconds
		s.Throttle = reply.Throttle
	}
	end := time.Now()
	log.Printf("Uploading took %s", end.Sub(start))
}

func (s *V4lStreamer) Update() {
	if s.Throttle != 0 {
		s.Throttle -= 1
		return
	}
	if (s.CanTick()) {
		s.SendFrame()
	}
}

func (s *V4lStreamer) SendFrame() {
	s.device.TurnOff()
	s.device.TurnOn(true)
	frame, t, err := s.CaptureFrame()
	if err != nil {
		log.Printf("Cannot capture frame %+v", err)
		return
	}
	s.writeUpdate(t, frame)
}

func (s *V4lStreamer) clearError() {
	s.lastErr = nil
}

func (s *V4lStreamer) CanTick() bool { return s.UploadsEnabled }
func (s *V4lStreamer) Tick() error {
	defer s.clearError()
	if (s.CanTick()) {
		s.SendFrame()
	}
	return s.lastErr
}
