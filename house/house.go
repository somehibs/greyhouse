package house

import (
	"git.circuitco.de/self/greyhouse/log"
	"math/rand"
	"time"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/presence"
	"git.circuitco.de/self/greyhouse/thirdparty"
)

type House struct {
	presence *presence.PresenceService
	rules *RuleService
	Rooms map[api.Room]Room
	leavingRoom map[api.Room]int
	enteringRoom map[api.Room]int
}

type Room struct {
	Lights []thirdparty.Light
}

func New(ruleService *RuleService, presenceService *presence.PresenceService) House {
	log.Print("Starting house...")
	house := House{Rooms: map[api.Room]Room{}, rules: ruleService, presence: presenceService, leavingRoom: map[api.Room]int{}, enteringRoom: map[api.Room]int{}}
	presenceService.AddCallback(house)
	//hueBridge := thirdparty.NewHueBridge("192.168.0.10")
	house.Rooms[api.Room_LOUNGE] = Room{
		Lights: []thirdparty.Light{
			//hueBridge.NewLight("lounge front"),
			//hueBridge.NewLight("lounge rear"),
		},
	}
	house.Rooms[api.Room_STUDY] = Room{
		Lights: []thirdparty.Light{
			//hueBridge.NewLight("study"),
		},
	}
	house.Rooms[api.Room_BEDROOM] = Room{
		Lights: []thirdparty.Light{
			//hueBridge.NewLight("bedroom overhead"),
		},
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
	return h.TryGetLightsImpl(room, false)
}

func (h House) TryGetLightsImpl(room api.Room, ignoreRules bool) []thirdparty.Light {
	if ignoreRules {
		return h.Rooms[room].Lights
	} else {
		rule := h.rules.ApplyRules(room)
		for _, modification := range rule {
			if modification.System == "LIGHT" && modification.Disable {
				log.Print("Disabling lights due to rule conditions.")
				return nil
			}
		}
		return h.Rooms[room].Lights
	}

}

func (h House) eventuallyEnterRoom(room api.Room, id int, pause int) {
	time.Sleep(time.Duration(pause)*time.Second)
	if h.enteringRoom[room] == id {
		lights := h.TryGetLights(room)
		log.Printf("wanna turn on %d lights in %+v", len(lights), room)
		if len(lights) > 0 {
			log.Printf("Turning on %d lights in %+v", len(lights), room)
			for _, light := range lights {
				light.On()
			}
		}
	}
}


func (h House) RoomPresenceChange(room api.Room, present int32) {
	if present > 0 {
		// turn on the lights
		identifier := rand.Int()
		h.leavingRoom[room] = 0
		h.enteringRoom[room] = identifier
		go h.eventuallyEnterRoom(room, identifier, 2)
	} else {
		// ignore leaving rooms
		identifier := rand.Int()
		h.enteringRoom[room] = 0
		h.leavingRoom[room] = identifier
		go h.eventuallyLeaveRoom(room, identifier, 60)
	}
}

func (h House) eventuallyLeaveRoom(room api.Room, id int, sleep int) {
	time.Sleep(time.Duration(sleep)*time.Second)
	if h.leavingRoom[room] == id {
		h.leavingRoom[room] = 0
		lights := h.TryGetLights(room)
		log.Printf("Turning off %d lights in %+v", len(lights), room)
		for _, light := range lights {
			light.Off()
		}
	}
}
