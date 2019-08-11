package recognise

// stubbed due to tf requirement
import (
	"log"
)

// Global labels array
type Recogniser struct {
}

func NewRecogniser(dataDir string) Recogniser {
	log.Print("Recogniser/TF is not supported on ARM")
	return Recogniser{}
}
func (r Recogniser) RecogniseImage(img []byte) []Object {
	return nil
}
