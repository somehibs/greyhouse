package main

import (
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/version"
)

var serverAddr = "sloth.local:9999"
var bindAddr = "0.0.0.0:9991"
var nodeIdentifier = "sloth"
var nodeRoom = api.Room_KITCHEN
var thisVersion = version.CurrentVersion()
var modules = make([]interface{}, 0)

func loadModules() {
	//modules = append(modules, Motion
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

		for ;; {
			log.Fatal("need some modules to process")
		}

		// Call Register with our favourite address
		i, e := nodeClient.Register(ctx, &api.NodeMetadata{
			Identifier: nodeIdentifier,
			ClientAddress: bindAddr,
			Room: nodeRoom,
			Version: &thisVersion,
		})
		log.Printf("R: %+v Error: %+v\n", i, e)
		time.Sleep(4*time.Second)
	}
}

