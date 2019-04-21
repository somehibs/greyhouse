package thirdparty

import (
	"time"
	"git.circuitco.de/self/greyhouse/log"
	api "git.circuitco.de/self/greyhouse/api"
)

type GoogleMapsLocationSharing struct {
	cookies map[string][]string
	lastRequest *time.Time
	locations map[PersonId]api.Location
}

func NewGoogleMapsLocationSharing() GoogleMapsLocationSharing {
	gmls := GoogleMapsLocationSharing{
		map[string][]string{},
		nil,
		map[PersonId]api.Location{},
	}
	return gmls
}

func (gmls *GoogleMapsLocationSharing) CacheOk() bool {
	return false
}

func (gmls *GoogleMapsLocationSharing) GetLocations() map[PersonId]api.Location {
	// make the api request, if it hasnt been 60s
	if !gmls.CacheOk() {
		err := gmls.refresh()
		if err != nil {
			log.Printf("Error retrieving locations: %+v", err)
		}
	}
	return gmls.locations
}

func (gmls *GoogleMapsLocationSharing) refresh() error {
	return nil
}
