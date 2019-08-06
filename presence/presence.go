package presence

import (
	"os"
	"time"
	"fmt"
	"encoding/csv"
	"git.circuitco.de/self/greyhouse/log"

	"golang.org/x/net/context"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/node"
	"git.circuitco.de/self/greyhouse/recognise"
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

	callbacks []PresenceCallback

	recognising map[api.Room]bool
	recognise recognise.Recogniser
}

func NewService(nodeService *node.NodeService) PresenceService {
	log.Print("Starting presence service...")
	recogniser := recognise.NewRecogniser("recognise/models")
	presence := PresenceService{
		nodeService,
		map[api.Room][]PresenceEvent{},
		make([]PresenceEvent, 0),
		make([]PresenceCallback, 0),
		map[api.Room]bool{},
		recogniser,
	}
	return presence
}

func (ps *PresenceService) AddCallback(callback PresenceCallback) {
	ps.callbacks = append(ps.callbacks, callback)
}

func (ps *PresenceService) RemoveCallback(removalCallback PresenceCallback) {
	for ind, callback := range ps.callbacks {
		if callback == removalCallback {
			ps.callbacks = append(ps.callbacks[:ind], ps.callbacks[:ind+1]...)
		}
	}
}

var throttle = int32(3)

func (ps *PresenceService) Image(ctx context.Context, update *api.ImageUpdate) (*api.PresenceUpdateReply, error) {
	node := ps.nodes.GetNode(ctx)
	go ps.recogniseImage(node, update)
	return &api.PresenceUpdateReply{Throttle: throttle}, nil
}

func (ps *PresenceService) recogniseImage(node *node.Node, update *api.ImageUpdate) {
	// Given a node, who is in this room?
	// If the room is black, or there's no matches, there's nothing to be done.
	if ps.recognising[node.Room] {
		log.Printf("Throttling room")
		return
	}
	ps.recognising[node.Room] = true
	start := time.Now()
	found := ps.recognise.RecogniseImage(update.Image)
	end := time.Now()
	diff := end.Sub(start)
	maybeThrottle := int32(float64(diff/time.Millisecond))/1000
	log.Printf("Maybe throttle: %d", maybeThrottle)
	log.Printf("Recognise time: %s", diff/time.Millisecond)
	log.Printf("Found: %+v", found)
	for _, match := range found {
		if match.Class == "person" || match.Class == "cat" {
			if match.Class == "person" {
				// Yep, someone's in here
			} else {
				// Stricter recognition requirements
			}
		}
		// who is in here?
	}
	ps.recognising[node.Room] = false
}

func (ps *PresenceService) Update(ctx context.Context, update *api.PresenceUpdate) (*api.PresenceUpdateReply, error) {
	unode := ps.nodes.GetNode(ctx)
	reply := &api.PresenceUpdateReply{Throttle: 0}
	event := PresenceEvent {
		*update,
		unode,
	}
	logUpdate(unode.Room, update)
	log.Debugf("New motion: %+v from node %+v", update.PeopleDetected, unode.Name)
	for _, callback := range ps.callbacks {
		callback.RoomPresenceChange(unode.Room, update.PeopleDetected)
	}
	switch update.Type {
		case api.PresenceType_Motion:
			ps.motionPresence[unode.Room] = append(ps.motionPresence[unode.Room], event)
			log.Printf("Motions recorded: %+v", len(ps.motionPresence))
			reply.Throttle = 15
	}
	return reply, nil
}

var MotionTimeFormat = "2006-01-02T15:04:05-0700"

func logUpdate(room api.Room, update *api.PresenceUpdate) error {
	statFile, err := os.Stat("motion.csv")
	f, err := os.OpenFile("motion.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	if statFile == nil || statFile.Size() == 0 {
		w.Write([]string{"time","room","source","state"})
	}
	w.Write([]string{time.Now().Format(MotionTimeFormat), fmt.Sprintf("%s", room), fmt.Sprintf("%s", update.Type), fmt.Sprintf("%d", update.PeopleDetected)})
	w.Flush()
	return nil
}

// We know when presences expire and when we need reprocessing.
// This duration might get longer, but if it's <5s, just call Tick()
func (ps PresenceService) NextTick() int32 {
	return 0
}

func (ps PresenceService) Tick() {
}
