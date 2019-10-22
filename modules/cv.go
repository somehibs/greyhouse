package modules

import (
	"bytes"
	hash "github.com/corona10/goimagehash"
	"image/jpeg"
	"log"
)

var (
	night_dhash = hash.NewImageHash(72340172838076673, hash.DHash)
)

type ComputerVision struct {
	video *V4lStreamer
	thresholds int
	deathThreshold float64
	lumenThresholdLow float64
	lumenThresholdHigh float64
	autoExposure bool
	frameChannel chan []byte
	lastDiffHash *hash.ImageHash
	largeExposureAdjust bool
}

func NewComputerVision(video *V4lStreamer) ComputerVision {
	return ComputerVision{video: video,
		lumenThresholdLow:  100.0,
		lumenThresholdHigh: 140.0,
		autoExposure:       true,
		frameChannel:       make(chan []byte),
	}
}

type NoVideoModuleError struct {
}

func (e NoVideoModuleError) Error() string {
	return "NoVideoModule"
}

func (cv *ComputerVision) Init(config ModuleConfig) error {
	if cv.video == nil {
		return NoVideoModuleError{}
	}
	if config.Args["Deathold"] != nil {
		cv.deathThreshold = (config.Args["Deathold"]).(float64)
	}
	if config.Args["Threshold"] != nil {
		cv.thresholds = int((config.Args["Threshold"]).(float64))
	}
	if config.Args["DisableExposure"] != nil {
		cv.autoExposure = !(config.Args["DisableExposure"]).(bool)
		log.Printf("Setting auto-exposure: %+v", cv.autoExposure)
	}
	if config.Args["MinLumen"] != nil {
		cv.lumenThresholdLow = (config.Args["MinLumen"]).(float64)
	}
	if config.Args["MaxLumen"] != nil {
		cv.lumenThresholdHigh = (config.Args["MaxLumen"]).(float64)
	}
	go cv.HandleFrames()
	cv.video.NewFrame(cv.frameChannel)
	return nil
}

func (cv *ComputerVision) SetDesiredExposure(exposure int32, adjustNow bool) {
	current := cv.video.GetExposureTime()
	if current == exposure {
		return
	}
	if adjustNow {
		cv.video.SetExposureTime(exposure)
		cv.largeExposureAdjust = true
		return
	}
	if current < exposure {
		distance := exposure - current
		inc := int32(10)
		if current < 200 {
			inc = 3
		} else if current < 500 {
			inc = 5
		} else if current > 1000 {
			inc = 25
		}
		if inc > distance {
			inc = distance
		}
		current += inc
	} else if current > exposure {
		distance := current - exposure
		sub := int32(10)
		if distance > 100 {
			sub = 30
		}
		if current > 5000 {
			sub = 100
		}
		current -= sub
	}
	cv.video.SetExposureTime(current)
}

func (cv *ComputerVision) HandleExposure(averageLumen float64) {
	if cv.autoExposure == false {
		return
	}
	if averageLumen < cv.lumenThresholdLow {
		cv.SetDesiredExposure(6000, averageLumen<cv.lumenThresholdLow-25)
	}
	if averageLumen > cv.lumenThresholdHigh {
		cv.SetDesiredExposure(0, false)
	}
}

var debug = false

func (cv *ComputerVision) HandleFrames() {
	skip := 0
	for {
		b := <-cv.frameChannel
		if skip > 0 {
			skip -= 1
			continue
		}
		// Read frame
		byteReader := bytes.NewReader(b)
		img, _ := jpeg.Decode(byteReader)

		// Hash and lumen frame
		diffHash, averageLumens, err := hash.DifferenceHash(img)
		if err != nil {
			continue
		} else if false {
			log.Printf("Hash type: %s (%d) hash: %d", diffHash.GetKind(), diffHash.GetKind(), diffHash.GetHash())
		}

		// Allow exposure changes based on lumens
		cv.HandleExposure(averageLumens)

		lightingThreshold := cv.getLightingThreshold(averageLumens)
		if cv.largeExposureAdjust {
			cv.largeExposureAdjust = false
			log.Printf("SUPPRESS TRIGGER")
			lightingThreshold = 100
		} else if averageLumens < cv.deathThreshold {
			if !debug {
				log.Printf("About to enter deep sleep...")
				debug = true
			} else {
				cv.video.DeepSleep()
			}
			continue
		} else if debug {
			debug = false
			log.Printf("Waking up...")
		}

		if cv.lastDiffHash != nil {
			d, _ := cv.lastDiffHash.Distance(diffHash)
			if d > 1+cv.thresholds+lightingThreshold {
				log.Printf("TRIGGER %d LUMEN %f", d, averageLumens)
			}
		}
		cv.lastDiffHash = diffHash
	}
}

func (cv *ComputerVision) CanTick() bool {
	return false
}

func (cv *ComputerVision) Tick() error {
	return nil
}

func (cv *ComputerVision) Update() {
}

func (cv *ComputerVision) Shutdown() {
}

func (cv *ComputerVision) getLightingThreshold(averageLumens float64) int {
	switch {
	case averageLumens < 5:
		return 7
	case averageLumens < 10:
		return 5
	case averageLumens < 15:
		return 3
	case averageLumens < 30:
		return 2
	case averageLumens < 40:
		return 1
	case averageLumens > 120:
		return 1
	default:
		return 0
	}
}
