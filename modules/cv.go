package modules

import (
	"log"
	"bytes"
	"image/jpeg"

	hash "github.com/corona10/goimagehash"
)

type ComputerVision struct {
	video *V4lStreamer
	frameChannel chan []byte
	lastHash *hash.ExtImageHash
}

func NewComputerVision(video *V4lStreamer) ComputerVision {
	return ComputerVision{video, make(chan []byte), nil}
}

func (cv *ComputerVision) Init(config ModuleConfig) error {
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
		if cv.lastHash != nil {
			d, _ := cv.lastHash.Distance(bhash)
			if d > 5 {
				log.Printf("Passed trigger distance with: %d", d)
			}
		}
		cv.lastHash = bhash
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
