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
	frameChannel chan []byte
	lastFrame []byte
	lastHash *hash.ExtImageHash
	lastDiffHash *hash.ImageHash
}

func NewComputerVision(video *V4lStreamer) ComputerVision {
	return ComputerVision{video, 0, false, make(chan []byte), nil, nil, nil}
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
	log.Print("CV starting")
	go cv.HandleFrames()
	cv.video.NewFrame(cv.frameChannel)
	return nil
}

func (cv *ComputerVision) HandleFrames() {
	for {
		b := <-cv.frameChannel
		byteReader := bytes.NewReader(b)
		img, _ := jpeg.Decode(byteReader)
		bhash, _ := hash.ExtAverageHash(img, 10, 10)
		dhash, _ := hash.DifferenceHash(img)
		night_time, _ := dhash.Distance(night_dhash)
		if night_time == 0 {
			// definitely night time
		} else if night_time < 3 {
			// very close to night time
		}
		if cv.lastHash != nil {
			d, _ := cv.lastHash.Distance(bhash)
			if d > cv.thresholds+10 {
				log.Printf("Average passed trigger: %d", d)
			}
		}
		if cv.lastDiffHash != nil {
			d, _ := cv.lastDiffHash.Distance(dhash)
			if d > cv.thresholds+3 {
				log.Printf("Difference passed trigger: %d", d)
			}
		}
		if cv.showHashes {
			log.Printf("Hash type: %s (%d) hash: %d", dhash.GetKind(), dhash.GetKind(), dhash.GetHash())
		}
		cv.lastFrame = b
		cv.lastHash = bhash
		cv.lastDiffHash = dhash
	}
}

func (cv *ComputerVision) CanTick() bool {
	return false
}

func (cv *ComputerVision) Tick() error {
	// Try to fetch an image
	return nil
}

func (cv *ComputerVision) Update() {
	// nothin
}

func (cv *ComputerVision) Shutdown() {
}
