package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/node"
	"git.circuitco.de/self/greyhouse/house"

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
		grpc.UnaryInterceptor(grpcm.ChainUnaryServer(util.LogInterceptor)),
		//grpc.UnaryInterceptor(util.LogInterceptor)
	)
	houseService := house.New()
	log.Printf("Made a house %s", houseService)
	nodeService := node.NewService()
	api.RegisterPrimaryNodeServer(server, nodeService)
	log.Print("Services listening forever.")
	server.Serve(listen)
	log.Fatal("Service is going down...")
}
