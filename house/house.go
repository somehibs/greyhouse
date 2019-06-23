package house

import (
	"git.circuitco.de/self/greyhouse/log"
	"time"

	"git.circuitco.de/self/greyhouse/thirdparty"
	"git.circuitco.de/self/greyhouse/presence"
	api "git.circuitco.de/self/greyhouse/api"
)

type House struct {
	presence *presence.PresenceService
	rules *RuleService
	Rooms map[api.Room]Room
}

type Room struct {
	Lights []thirdparty.Light
}

func New(ruleService *RuleService, presenceService *presence.PresenceService) House {
	log.Print("Starting house...")
	house := House{Rooms: map[api.Room]Room{}, rules: ruleService, presence: presenceService}
	presenceService.AddCallback(house)
	hueBridge := thirdparty.NewHueBridge("192.168.0.17")
	house.Rooms[api.Room_LOUNGE] = Room{
		Lights: []thirdparty.Light{
			//hueBridge.NewLight("lounge front"),
			//hueBridge.NewLight("lounge rear"),
		},
	}
	house.Rooms[api.Room_STUDY] = Room{
		Lights: []thirdparty.Light{
			hueBridge.NewLight("study"),
		},
	}
	house.Rooms[api.Room_BEDROOM] = Room{
		Lights: []thirdparty.Light{},
//			hueBridge.NewLight("bedroom")
//		},
	}
	return house
}

func (h House) StartTicking() {
	go func() {
		var i = int64(0)
		for ;; {
			if i % 60 == 0 {
				// Approximately 1m
				h.TickMinute()
			}
			h.Tick()
			i += 1
			time.Sleep(1*time.Second)
		}
	}()
}

func (h House) Tick() {
}

func (h House) TickMinute() {
}

func (h House) PersonLocationUpdate(personId int64) {
}
func (h House) TryGetLights(room api.Room) []thirdparty.Light {
	return h.TryGetLightsImpl(room, true)
}

func (h House) TryGetLightsImpl(room api.Room, ignoreRules bool) []thirdparty.Light {
	if ignoreRules {
		return h.Rooms[room].Lights
	} else {
		// ask rules if room is restricted from lights
		//
		return nil
	}

}

func (h House) RoomPresenceChange(room api.Room, present int32) {
	if present > 0 {
		// turn on the lights
		log.Printf("Turning on %d lights in %+v", len(h.TryGetLights(room)), room)
		for _, light := range h.TryGetLights(room) {
			light.On()
		}
	} else {
		// ignore leaving rooms for now
	}
}
