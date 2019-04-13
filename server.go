package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/node"
	"git.circuitco.de/self/greyhouse/house"
	"git.circuitco.de/self/greyhouse/presence"

	util "git.circuitco.de/self/grpc-util"

	grpcm "github.com/grpc-ecosystem/go-grpc-middleware"
)

var bindAddr = "0.0.0.0:9999"

func main() {
	listen, err := net.Listen("tcp", bindAddr)
	if err != nil {
		log.Fatalf("Failed to listen to %s because %s", bindAddr, err)
	}

	log.Print("Starting service.")
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcm.ChainUnaryServer(util.LogInterceptor, node.NodeInterceptor)),
		//grpc.UnaryInterceptor(util.LogInterceptor)
	)
	node.AllowedMethods["/greymatter/node.Register"] = true
	
	rulesService := house.NewRuleService()
	api.RegisterRulesServer(server, rulesService)

	houseService := house.New(rulesService)
	log.Printf("Made a house %s", houseService)

	nodeService := node.NewService()
	api.RegisterPrimaryNodeServer(server, nodeService)

	personService := presence.NewPersonService()
	api.RegisterPersonServer(server, personService)

	presenceService := presence.NewService()
	api.RegisterPresenceServer(server, presenceService)

	log.Print("Services listening forever.")
	server.Serve(listen)
	log.Fatal("Service is going down...")
}
