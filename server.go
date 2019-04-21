package main

import (
	"log"
	"net"
	"time"
	"math/rand"

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
	rand.Seed(time.Now().UnixNano())
	listen, err := net.Listen("tcp", bindAddr)
	if err != nil {
		log.Fatalf("Failed to listen to %s because %s", bindAddr, err)
	}

	log.Print("Starting service.")
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcm.ChainUnaryServer(util.LogInterceptor, util.AuthenticationInterceptor)),
		//grpc.UnaryInterceptor(util.LogInterceptor)
	)
	util.NoAuthMethods["/greyhouse.PrimaryNode/Register"] = true

	rulesService := house.NewRuleService()
	api.RegisterRulesServer(server, rulesService)

	nodeService := node.NewService()
	util.AuthChecker = nodeService
	api.RegisterPrimaryNodeServer(server, nodeService)

	personService := presence.NewPersonService()
	api.RegisterPersonServer(server, personService)

	presenceService := presence.NewService(&nodeService)
	api.RegisterPresenceServer(server, &presenceService)

	houseService := house.New(&rulesService, &presenceService)
	log.Printf("Made a house %s", houseService)

	log.Print("Starting house tick thread.")
	houseService.StartTicking()

	log.Print("Services listening now.")
	server.Serve(listen)
	log.Fatal("Service is going down...")
}
