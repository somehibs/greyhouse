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
	callbacks []chan<- []byte
	device *v4l.Device
	framesCaught int
	devicePath string
	lastFrame []byte
	lastErr error
	lastUpload *time.Time
	Throttle int32
	UploadsEnabled bool
}

func NewV4lStreamer() V4lStreamer {
	return V4lStreamer{make([]chan<- []byte, 0), nil, 0, "", nil, nil, nil, 0, true}
}

func (s *V4lStreamer) NewFrame(listener chan<- []byte) {
	s.callbacks = append(s.callbacks, listener)
}

func (s *V4lStreamer) StopFrame(listener chan<- []byte) {
	log.Print("StopFrame not supported")
}

func (s *V4lStreamer) restarter() {
	for {
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
		s.devicePath = deviceInfo.Path
		err := s.OpenDevice()
		if err != nil {
			return err
		}
		err = s.ConfigDevice()
		if err != nil {
			return err
		}
		s.device.TurnOn(false)
		s.device.TurnOff()
		s.listenHttp()
		return err
	} else {
		log.Printf("Could not find any devices for streamer.")
		return errors.New("v4l_no_devices")
	}
}

func (s *V4lStreamer) OpenDevice() error {
	log.Print("Opening video device at path " + s.devicePath)
	if s.device != nil {
		s.device.Close()
	}
	device, err := v4l.Open(s.devicePath)
	if err != nil {
		log.Printf("Err opening device: %+v", err)
		return err
	}
	s.device = device
	return nil
}

func (s *V4lStreamer) ConfigDevice() error {
	// manually set the device config to what we want
	err := s.device.SetConfig(v4l.DeviceConfig{Width: 640, Height: 480, Format: FourCC([]byte{'M','J','P','G'}), FPS: v4l.Frac{15, 1}})
	if err != nil {
		return err
	}
	gcf, err := s.device.GetConfig()
	if err != nil {
		return err
	}
	log.Printf("gcf: %+v", gcf)
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
		w.Write(s.lastFrame)
	})
	go server.ListenAndServe()
}

func (s *V4lStreamer) DispatchFrame() {
	for _, c := range s.callbacks {
		c <- s.lastFrame
	}
}

func (s *V4lStreamer) CaptureFrame() ([]byte, time.Time, error) {
	s.framesCaught += 1
	if s.framesCaught % 500 == 0 {
		s.framesCaught = 0
		err := s.OpenDevice()
		if err != nil {
			return nil, time.Now(), err
		}
	}
	start := time.Now()
	s.device.TurnOn(true)
	buffer, err := s.device.Capture()
	var bufferCopy []byte
	if err == nil {
		bufferCopy = make([]byte, len(buffer.Source()))
		copy(bufferCopy, buffer.Source())
		s.lastFrame = bufferCopy
		go s.DispatchFrame()
	} else {
		log.Printf("Could not read frame: %s", err.Error())
	}
	s.device.TurnOff()
	end := time.Now()
	if false {
		log.Printf("Capture time %+v", end.Sub(start))
	}
	return bufferCopy, start, err
}

func (s *V4lStreamer) Shutdown() {
	log.Print("streamer shutting down")
	if s.device != nil {
		s.device.Close()
	}
}

func (s *V4lStreamer) writeUpdate(t time.Time, img []byte) {
	if chost == nil {
		log.Print("Cannot report image frame to empty chost")
		return
	}
	ctx := (*chost).GetContext()
	update := api.ImageUpdate {
		Time: t.Unix(),
		Image: img,
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
	if false {
		log.Printf("Uploading took %s", end.Sub(start))
	}
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
