package main

import (
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"

	api "git.circuitco.de/self/greyhouse/api"
	// this was a mistake, should have single-packaged everything that would have fit in one package
	// TODO: repackage node, house, presence and web as greyhouse
	"git.circuitco.de/self/greyhouse/house"
	"git.circuitco.de/self/greyhouse/node"
	"git.circuitco.de/self/greyhouse/presence"
	"git.circuitco.de/self/greyhouse/web"

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

	log.Print("Starting public webserver")
	web.Route(":9998", &nodeService)

	log.Print("Starting house tick thread.")
	houseService.StartTicking()

	log.Print("Services listening now.")

	log.Fatalf("Service is going down... (e: %s)", server.Serve(listen))
}
