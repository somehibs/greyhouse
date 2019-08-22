package modules

import (
	"log"
	"os"
	"bytes"
	"image/jpeg"

	hash "github.com/corona10/goimagehash"
)

type ComputerVision struct {
	video *V4lStreamer
	thresholds int
	frameChannel chan []byte
	lastFrame []byte
	lastHash *hash.ExtImageHash
	lastDiffHash *hash.ImageHash
}

func NewComputerVision(video *V4lStreamer) ComputerVision {
	return ComputerVision{video, 0, make(chan []byte), nil, nil, nil}
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
		triggerPassed := 0
		if cv.lastHash != nil {
			d, _ := cv.lastHash.Distance(bhash)
			if d > cv.thresholds+10 {
				triggerPassed = d
				log.Printf("Average passed trigger: %d", d)
			}
		}
		if cv.lastDiffHash != nil {
			d, _ := cv.lastDiffHash.Distance(dhash)
			if d > cv.thresholds+3 {
				triggerPassed = d
				log.Printf("Difference passed trigger: %d", d)
			}
		}
		if triggerPassed > 0 {
			log.Println("writing file to disk because trigger")
			f, err := os.Create("trigger.jpg")
			if err == nil {
				f.Write(b)
				f.Close()
			}
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
