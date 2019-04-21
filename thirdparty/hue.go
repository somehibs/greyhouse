package thirdparty

import (
	"log"
	"strings"

	"github.com/amimof/huego"
)

type HueBridge struct {
	Address string
	bridge *huego.Bridge
}

type HueLight struct {
	light *huego.Light
}

var authorizePhase = true
var disable = false
var hueUser = "greyhouse"
var hueKey = ""

func NewHueBridge(addr string) HueBridge {
	bridge := HueBridge{Address: addr, bridge: huego.New(addr, hueKey)}
	if disable {
		return bridge
	}
	if len(hueKey) == 0 {
		discovered, err := huego.Discover()
		if err != nil {
			log.Fatalf("Cannot discover huego: %s", err.Error())
		}
		user, err := discovered.CreateUser(hueUser)
		if err != nil {
			log.Fatalf("Cannot create user: %s", err.Error())
		}
		log.Printf("User created: %s please note this down for future runs", user)
		bridge.bridge = discovered.Login(user)
	}
	return bridge
}

func (h HueBridge) getLight(lightName string) *huego.Light {
	lights, e := h.bridge.GetLights()
	if e != nil {
		log.Fatalf("Couldn't fetch lights from bridge: %s", e.Error())
	}
	for _, light := range lights {
		if strings.Compare(light.Name, lightName) == 0 {
			return &light
		}
	}
	log.Fatalf("Failed to find light %s", lightName)
	return nil
}

func (h HueBridge) NewLight(lightName string) (Light) {
	if disable {
		return HueLight{}
	}
	return HueLight{h.getLight(lightName)}
}

func (l HueLight) Brightness(bri int32) error {
	return l.light.Bri(uint8(bri))
}

func (l HueLight) Flash() error {
	return l.light.Alert("select")
}
func (l HueLight) Off() error {
	return l.light.Off()
}
func (l HueLight) On() error {
	return l.light.On()
}
