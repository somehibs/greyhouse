package presence

import (
	"time"
	"log"

	"golang.org/x/net/context"

	api "git.circuitco.de/self/greyhouse/api"
)

type PresenceEvent struct {
	sensor string
	presenceType api.PresenceType
	peopleDetected int32
	knownPeopleDetected []*Person
	lastSeen time.Time
	room api.Room
}

type PresenceCallback interface {
	// The user's location can be accurately judged, so we're giving you an update.
	// A lat/lon of 0/0 indicates the person doesn't have a location, so assume they're in the house.
	PersonLocationUpdate(int64)
	// Indicates the number of occupants in a room has changed.
	// house can inspect the room and use Person.Location
	// Determining this will be a lot easier when doors have entrance sensors wired close to the floor.
	RoomPresenceChange(api.Room, int32)
}

type PresenceService struct {
	presences []PresenceEvent
	// Accepts data from Presence modules.
	// Presence modules should do their best to provide as much useful data about what presence they detect.
	// Consider wide angle lens Pi cameras to detect motion
	// Some of the modules aren't associated with any room, but can help to indicate a room's presence (e.g. wifi indicator)
}

func NewService() PresenceService {
	log.Print("Starting presence service...")
	presence := PresenceService{make([]PresenceEvent, 0)}
	return presence
}

func (ps PresenceService) Update(ctx context.Context, presenceUpdate *api.PresenceUpdate) (*api.PresenceUpdateReply, error) {
	//ps.presences = append(ps.presences, presenceUpdate)
	reply := &api.PresenceUpdateReply{Throttle: 0}
	if presenceUpdate.Type == api.PresenceType_Motion {
		reply.Throttle = 15
	}
	return reply, nil
}
