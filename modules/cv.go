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
	showHashes bool
	autoExposure bool
	frameChannel chan []byte
	lastFrame []byte
	lastDiffHash *hash.ImageHash
}

func NewComputerVision(video *V4lStreamer) ComputerVision {
	return ComputerVision{video, 0, false, true, make(chan []byte), nil, nil}
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
	go cv.HandleFrames()
	cv.video.NewFrame(cv.frameChannel)
	return nil
}

func (cv *ComputerVision) HandleExposure(dhash *hash.ImageHash, averageLumen float64) bool {
	if cv.autoExposure == false {
		return false
	}
	// See if we match darkness hash
	//night_time, _ := dhash.Distance(night_dhash)
	if averageLumen < 2 {
		cv.video.SetExposureTime(10000)
		return true
	}
	log.Printf("Lumens: %f", averageLumen)
	return false
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
		if cv.lastDiffHash != nil {
			d, _ := cv.lastDiffHash.Distance(dhash)
			if d > cv.thresholds+3 {
				log.Printf("Difference passed trigger: %d", d)
			}
		}
		if cv.showHashes {
			log.Printf("Hash type: %s (%d) hash: %d", dhash.GetKind(), dhash.GetKind(), dhash.GetHash())
		}
		night := cv.HandleExposure(dhash, avg)
		if night {
			// automatically skip the next 5 frames when nighttime
			skip = 5
		}
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
