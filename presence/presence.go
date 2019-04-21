package presence

import (
	"time"
	"log"

	"golang.org/x/net/context"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/node"
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
	nodes *node.NodeService
	motionPresence map[api.Room]PresenceEvent

	// Presences based on phones. WiFi, GPS and other services might be utilised.
	phonePresence []PresenceEvent
}

func NewService(nodeService *node.NodeService) PresenceService {
	log.Print("Starting presence service...")
	presence := PresenceService{
		nodeService,
		map[api.Room]PresenceEvent{},
		make([]PresenceEvent, 0),
	}
	return presence
}

func (ps PresenceService) Update(ctx context.Context, update *api.PresenceUpdate) (*api.PresenceUpdateReply, error) {
	node := nodes.GetNode(ctx)
	switch update.Type {
		case api.PresenceType_Motion:
			ps.motionPresence[node.Room] = append(ps.motionPresence[node.Room], update)
			log.Printf("Motions recorded: %+v", len(ps.motionPresence))
	}
	reply := &api.PresenceUpdateReply{Throttle: 0}
	if presenceUpdate.Type == api.PresenceType_Motion {
		reply.Throttle = 15
	}
	return reply, nil
}
