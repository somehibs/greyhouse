package house

import (
	//"log"

	"git.circuitco.de/self/greyhouse/thirdparty"
	api "git.circuitco.de/self/greyhouse/api"
)

type House struct {
	Rooms map[api.Room]Room
}

type Room struct {
	Lights []thirdparty.Light
}

func New() House {
	house := House{Rooms: map[api.Room]Room{}}
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
