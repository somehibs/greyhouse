package presence

import (
	"git.circuitco.de/self/greyhouse/log"

	"golang.org/x/net/context"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/node"
)

type PresenceCallback interface {
	// The user's location can be accurately judged, so we're giving you an update.
	// A lat/lon of 0/0 indicates the person doesn't have a location. assume they're in the house.
	PersonLocationUpdate(int64)
	// Indicates the number of occupants in a room has changed.
	RoomPresenceChange(api.Room, int32)
}

type PresenceEvent struct {
	update api.PresenceUpdate
	node *node.Node
}

type PresenceService struct {
	nodes *node.NodeService

	// Currently just PIR
	motionPresence map[api.Room][]PresenceEvent

	// Presences based on phones. WiFi, GPS and other services might be utilised.
	// Sometimes very accurate, sometimes useless.
	phonePresence []PresenceEvent
}

func NewService(nodeService *node.NodeService) PresenceService {
	log.Print("Starting presence service...")
	presence := PresenceService{
		nodeService,
		map[api.Room][]PresenceEvent{},
		make([]PresenceEvent, 0),
	}
	return presence
}

func (ps PresenceService) Update(ctx context.Context, update *api.PresenceUpdate) (*api.PresenceUpdateReply, error) {
	unode := ps.nodes.GetNode(ctx)
	reply := &api.PresenceUpdateReply{Throttle: 0}
	event := PresenceEvent {
		*update,
		unode,
	}
	log.Debugf("New motion: %+v", event)
	switch update.Type {
		case api.PresenceType_Motion:
			ps.motionPresence[unode.Room] = append(ps.motionPresence[unode.Room], event)
			log.Printf("Motions recorded: %+v", len(ps.motionPresence))
			reply.Throttle = 15
	}
	return reply, nil
}

// We know when presences expire and when we need reprocessing.
// This duration might get longer, but if it's <5s, just call Tick()
func (ps PresenceService) NextTick() int32 {
	return 0
}

func (ps PresenceService) Tick() {
}
