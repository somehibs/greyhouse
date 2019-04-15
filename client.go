package main

import (
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/version"

	"git.circuitco.de/self/greyhouse/modules"
)

var serverAddr = "sloth.local:9999"
var bindAddr = "0.0.0.0:9991"
var nodeIdentifier = "sloth"
var nodeRoom = api.Room_KITCHEN
var thisVersion = version.CurrentVersion()
var loadedModules = make([]modules.GreyhouseClientModule, 0)
var tickModules = make([]modules.GreyhouseClientModule, 0)

func loadModules() {
	// At the moment, this is just manual based on configuration that's not written yet
	loadedModules = append(loadedModules, modules.NewGpioWatcher(2))
	for _, module := range loadedModules {
		module.Init()
		if module.CanTick() {
			tickModules = append(tickModules, module)
		}
	}
}

func registered(clientHost modules.ClientHost) {
	// refresh the modules
	for _, module := range loadedModules {
		module.Update(&clientHost)
	}
}

func getClients(conn *grpc.ClientConn, nodeClient *api.PrimaryNodeClient, nodeKey string) modules.ClientHost {
	ch := modules.ClientHost{Node: nodeClient}
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
			time.Sleep(100*time.Second)
		} else {
			log.Printf("R: %+v Error: %+v\n", i, e)
		}
		time.Sleep(4*time.Second)
	}
}

