package presence

import (
	"log"

	"golang.org/x/net/context"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/node"
)

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
	motionPresence map[api.Room][]api.PresenceUpdate

	// Presences based on phones. WiFi, GPS and other services might be utilised.
	phonePresence []api.PresenceUpdate
}

func NewService(nodeService *node.NodeService) PresenceService {
	log.Print("Starting presence service...")
	presence := PresenceService{
		nodeService,
		map[api.Room][]api.PresenceUpdate{},
		make([]api.PresenceUpdate, 0),
	}
	return presence
}

func (ps PresenceService) Update(ctx context.Context, update *api.PresenceUpdate) (*api.PresenceUpdateReply, error) {
	unode := ps.nodes.GetNode(ctx)
	reply := &api.PresenceUpdateReply{Throttle: 0}
	switch update.Type {
		case api.PresenceType_Motion:
			ps.motionPresence[unode.Room] = append(ps.motionPresence[unode.Room], *update)
			log.Printf("Motions recorded: %+v", len(ps.motionPresence))
			reply.Throttle = 15
	}
	return reply, nil
}
