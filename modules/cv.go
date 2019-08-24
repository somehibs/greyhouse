package modules

import (
	"log"
	"bytes"
	"image/jpeg"

	hash "github.com/corona10/goimagehash"
)

// what sort of things should this module report?
// well, it's computer vision, so it should tell the system about any motion change
// it should tell the system the magnitude of motion

// since it's cv, we should look to recognise an empty room
// if possible, we will try obtaining an outline of the empty room
// we can obtain an outline of the room's default state
// then we can run difference hashes on future outlines to see if there's a significant change in outline

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
	return ComputerVision{video, 0, 100.0, 140.0, false, true, make(chan []byte), nil, nil, false}
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
		log.Print("Showing hashes: %+v", cv.showHashes)
	}
	if config.Args["DisableExposure"] != nil {
		cv.autoExposure = !(config.Args["DisableExposure"]).(bool)
		log.Print("Showing hashes: %+v", cv.showHashes)
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
	log.Printf("Lumens: %f", averageLumen)
	if averageLumen < cv.lumenThresholdLow {
		cv.SetDesiredExposure(6000, averageLumen<cv.lumenThresholdLow-20)
	}
	if averageLumen > cv.lumenThresholdHigh {
		log.Printf("desired exposure: %f", averageLumen)
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
		dhash, avg, _ := hash.DifferenceHash(img)
		extra_threshold := 0
		//if avg < 3 {
			// automatically skip the next 5 frames when nighttime
			//skip = 5
		//} else
		// very low lighting, unreliable
		if avg < 18 {
			extra_threshold = 6
		// prone to some banding
		} else if avg < 30 {
			extra_threshold = 2
		} else if avg < 40 {
			extra_threshold = 1
		}
		if cv.largeExposureAdjust {
			cv.largeExposureAdjust = false
			extra_threshold = 100
			log.Printf("SUPPRESS TRIGGER")
		}
		if cv.lastDiffHash != nil {
			d, _ := cv.lastDiffHash.Distance(dhash)
			if d > 1+cv.thresholds+extra_threshold {
				log.Printf("Difference passed trigger: %d", d)
			}
		}
		if cv.showHashes {
			log.Printf("Hash type: %s (%d) hash: %d", dhash.GetKind(), dhash.GetKind(), dhash.GetHash())
		}
		cv.HandleExposure(dhash, avg)
		cv.lastFrame = b
		cv.lastDiffHash = dhash
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
