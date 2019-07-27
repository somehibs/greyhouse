package main

import (
	"log"
	"time"
	"os"
	"errors"
	"os/signal"
	"syscall"
	"encoding/json"
	"io/ioutil"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/version"

	"git.circuitco.de/self/greyhouse/modules"
)

type ModuleConfig struct {
	Name string
	Args []string
}

type ClientConfig struct {
	Node string
	NodeAddress string
	Room api.Room
	Server string
	Modules []ModuleConfig
}

func loadClientConfig() (ClientConfig, error) {
	clientConfig := ClientConfig{}
	f, err := os.Open("client.json")
	if err != nil {
		return clientConfig, err
	}
	read, err := ioutil.ReadAll(f)
	if err != nil {
		return clientConfig, err
	}
	err = json.Unmarshal(read, &clientConfig)
	if clientConfig.Room == 0 {
		return clientConfig, errors.New("Room not correctly set.")
	}
	return clientConfig, err

}

var bindAddr = "0.0.0.0:9991" // not implemented
var thisVersion = version.CurrentVersion()
var loadedModules = make([]modules.GreyhouseClientModule, 0)
var tickModules = make([]modules.GreyhouseClientModule, 0)

func loadModules(moduleConfig []ModuleConfig) {
	log.Print("loading modules")
	for _, config := range moduleConfig {
		var module modules.GreyhouseClientModule
		switch config.Name {
		case "gpio":
			gpio := modules.NewGpioWatcher(23)
			module = &gpio
		default:
			log.Panicf("module name not recognised: %+v\n", config)
		}
		if module != nil {
			loadedModules = append(loadedModules, module)
		}
	}
	for _, module := range loadedModules {
		e := module.Init()
		if e != nil {
			log.Fatalf("couldnt load a module: %+v", e)
		}
		if module.CanTick() {
			tickModules = append(tickModules, module)
		}
	}
	shutdownSignal()
	log.Print("loaded")
}

func shutdownSignal() {
	// trap signals
	log.Print("trapping shutdown to allow for module shutdown...")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		for _, module := range loadedModules {
			module.Shutdown()
		}
		panic("Interrupted.")
	}()
}

func registered(clientHost modules.ClientHost) {
	// refresh the modules
	for _, module := range loadedModules {
		module.Update(&clientHost)
	}
	// spin on ticking unless an error comes back about networking
	tickCount := 0
	for ;; {
		for _, module := range tickModules {
			e := module.Tick()
			if e != nil {
				log.Print("FATAL Could not tick module: %+v due to %+v", module, e)
			}
		}
		time.Sleep(1*time.Second)
		tickCount += 1
		if tickCount % 10 == 0 {
			for _, module := range loadedModules {
				e := module.Tick()
				if e != nil {
					log.Printf("all modules tick found %s in %+v", e.Error(), module)
					return
				}
			}
		}
	}
}

func getClients(conn *grpc.ClientConn, nodeClient *api.PrimaryNodeClient, nodeKey string) modules.ClientHost {
	ch := modules.ClientHost{Node: nodeClient, Key: nodeKey}
	pr := api.NewPresenceClient(conn)
	// golang, because we've still got the same constructs as c and can't take the addr of a return value
	ch.Presence = &pr
	pe := api.NewPersonClient(conn)
	ch.Person = &pe
	r := api.NewRulesClient(conn)
	ch.Rules = &r
	return ch
}

func main() {
	log.Print("started")
	log.Print("loading config...")
	config, err := loadClientConfig()
	if err != nil {
		panic("Could not load config json: " + err.Error())
	}
	loadModules(config.Modules)

	for ;; {
		conn, err := grpc.Dial(config.Server, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect to %s: %s", config.Server, err.Error())
		}

		nodeClient := api.NewPrimaryNodeClient(conn)
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		//for ;; {
		//	log.Fatal("need some modules to process")
		//}

		// Call Register with our favourite address
		i, e := nodeClient.Register(ctx, &api.NodeMetadata{
			Identifier: config.Node,
			ClientAddress: config.NodeAddress,
			Room: config.Room,
			Version: &thisVersion,
		})
		if e == nil {
			// Perfect, we connected ok
			rc := api.NewRulesClient(conn)
			clientHost := getClients(conn, &nodeClient, i.Key)
			l, err := rc.List(clientHost.GetContext(), &api.RuleFilter{})
			if err != nil {
				log.Print("Could not get list: " + err.Error())
			} else {
				log.Printf("Rule list: %+v\n", l)
			}
			registered(clientHost)
			log.Print("Warning: registered() returned, authentication failure? retry connection")
		} else {
			log.Printf("R: %+v Error: %+v\n", i, e)
		}
		time.Sleep(4*time.Second)
	}
}

