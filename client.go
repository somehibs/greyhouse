package main

import (
	"log"
	"time"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/version"

	"git.circuitco.de/self/greyhouse/modules"
)

var serverAddr = "sloth.local:9999"
var bindAddr = "0.0.0.0:9991" // not implemented
var nodeIdentifier = "sloth"
var nodeRoom = api.Room_KITCHEN
var thisVersion = version.CurrentVersion()
var loadedModules = make([]modules.GreyhouseClientModule, 0)
var tickModules = make([]modules.GreyhouseClientModule, 0)

func loadModules() {
	// At the moment, this is just manual based on configuration that's not written yet
	log.Print("loading modules")
	loadedModules = append(loadedModules, modules.NewGpioWatcher(23))
	for _, module := range loadedModules {
		e := module.Init()
		if e != nil {
			log.Fatalf("couldnt load module: %+v", e)
		}
		if module.CanTick() {
			tickModules = append(tickModules, module)
		}
	}
	log.Print("modules loaded")
}

func registered(clientHost modules.ClientHost) {
	// refresh the modules
	for _, module := range loadedModules {
		module.Update(&clientHost)
	}
	// trap signals
	log.Print("Trapping signals")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		for _, module := range loadedModules {
			module.Shutdown()
		}
	}()
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
					log.Print("all modules tick found %+v in %+v", e, module)
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
	loadModules()

	for ;; {
		conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect to service %s because %s", serverAddr, err.Error())
		}

		nodeClient := api.NewPrimaryNodeClient(conn)
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		//for ;; {
		//	log.Fatal("need some modules to process")
		//}

		// Call Register with our favourite address
		i, e := nodeClient.Register(ctx, &api.NodeMetadata{
			Identifier: nodeIdentifier,
			ClientAddress: bindAddr,
			Room: nodeRoom,
			Version: &thisVersion,
		})
		if e == nil {
			// Perfect, we connected ok
			clientHost := getClients(conn, &nodeClient, i.Key)
			registered(clientHost)
		} else {
			log.Printf("R: %+v Error: %+v\n", i, e)
		}
		time.Sleep(4*time.Second)
	}
}

