package house

import (
	"log"
	"time"

	"git.circuitco.de/self/greyhouse/thirdparty"
	api "git.circuitco.de/self/greyhouse/api"
)

type House struct {
	rules RuleService
	Rooms map[api.Room]Room
}

type Room struct {
	Lights []thirdparty.Light
}

func New(ruleService RuleService) House {
	log.Print("Starting house...")
	house := House{Rooms: map[api.Room]Room{}, rules: ruleService}
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
		for ;; {
			h.Tick()
			time.Sleep(1*time.Second)
		}
	}()
}

func (h House) Tick() {
	log.Print("Tick.")
}
