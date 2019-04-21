package house

import (
	"log"
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
	hueBridge := thirdparty.NewHueBridge("192.168.0.17")
	house.Rooms[api.Room_LOUNGE] = Room{
		Lights: []thirdparty.Light{
			hueBridge.NewLight("lounge front"),
			hueBridge.NewLight("lounge rear"),
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
	// We have advanced a second.
}
