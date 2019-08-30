package modules

import (
	"log"
	"bytes"
	"image/jpeg"

	hash "github.com/corona10/goimagehash"
)

var (
	night_dhash = hash.NewImageHash(72340172838076673, hash.DHash)
)

type ComputerVision struct {
	video *V4lStreamer
	thresholds int
	lumenThresholdLow float64
	lumenThresholdHigh float64
	showHashes bool
	autoExposure bool
	frameChannel chan []byte
	lastFrame []byte
	lastDiffHash *hash.ImageHash
	largeExposureAdjust bool
}

func NewComputerVision(video *V4lStreamer) ComputerVision {
	return ComputerVision{video,
		0,
		100.0,
		140.0,
		false,
		true,
		make(chan []byte),
		nil,
		nil,
		false,
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
	if config.Args["Threshold"] != nil {
		cv.thresholds = int((config.Args["Threshold"]).(float64))
	}
	if config.Args["ShowHashes"] != nil {
		cv.showHashes = (config.Args["ShowHashes"]).(bool)
		log.Printf("Showing hashes: %+v", cv.showHashes)
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

func (cv *ComputerVision) HandleExposure(dhash *hash.ImageHash, averageLumen float64) {
	if cv.autoExposure == false {
		return
	}
	// See if we match darkness hash
	//night_time, _ := dhash.Distance(night_dhash)
	if averageLumen < cv.lumenThresholdLow {
		cv.SetDesiredExposure(6000, averageLumen<cv.lumenThresholdLow-20)
	}
	if averageLumen > cv.lumenThresholdHigh {
		// will lower by 10 until 0
		cv.SetDesiredExposure(0, false)
	}
}

func (cv *ComputerVision) HandleFrames() {
	skip := 0
	for {
		b := <-cv.frameChannel
		if skip > 0 {
			skip -= 1
			continue
		}
		byteReader := bytes.NewReader(b)
		img, _ := jpeg.Decode(byteReader)
		diffHash, avg, _ := hash.DifferenceHash(img)
		extraThreshold := 0
		//if avg < 3 {
			// automatically skip the next 5 frames when nighttime
			//skip = 5
		//} else
		// very low lighting, unreliable
		if avg < 10 {
			extraThreshold = 5
		} else if avg < 15 {
			extraThreshold = 3
		// prone to some banding
		} else if avg < 30 {
			extraThreshold = 2
		} else if avg < 40 {
			extraThreshold = 1
		}
		if cv.largeExposureAdjust {
			cv.largeExposureAdjust = false
			extraThreshold = 100
			log.Printf("SUPPRESS TRIGGER")
		}
		if cv.lastDiffHash != nil {
			d, _ := cv.lastDiffHash.Distance(diffHash)
			if d > 1+cv.thresholds+extraThreshold {
				log.Printf("TRIGGER %d LUMEN %f", d, avg)
			}
		}
		if cv.showHashes {
			log.Printf("Hash type: %s (%d) hash: %d", diffHash.GetKind(), diffHash.GetKind(), diffHash.GetHash())
		}
		cv.HandleExposure(diffHash, avg)
		cv.lastFrame = b
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
