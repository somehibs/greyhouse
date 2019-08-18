package modules

import (
	"log"

	"gocv.io/x/gocv/contrib"
)

type ComputerVision struct {
	streamer *V4lStreamer
	//hashes []contrib.ImgHashBase
}

func NewComputerVision(video *V4lStreamer) ComputerVision {
	return ComputerVision{video}//, make([]contrib.ImgHashBase, 0)}
}

func (cv *ComputerVision) Init(config ModuleConfig) error {
	log.Print("Starting CV")
	contrib.PHash{}
	hashes = append(hashes, contrib.AverageHash{})
	hashes = append(hashes, contrib.BlockMeanHash{})
	hashes = append(hashes, contrib.BlockMeanHash{Mode: contrib.BlockMeanHashMode1})
	hashes = append(hashes, contrib.ColorMomentHash{})
	hashes = append(hashes, contrib.NewMarrHildrethHash())
	hashes = append(hashes, contrib.NewRadialVarianceHash())
	return nil
}

func (cv *ComputerVision) CanTick() bool {
	return false
}

func (cv *ComputerVision) Tick() error {
	return nil
}

func (cv *ComputerVision) Update() {
	// nothin
}

func (cv *ComputerVision) Shutdown() {
}
