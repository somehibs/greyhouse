package main

import (
	"log"
	"net"

	"google.golang.orc/grpc"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/node"
	"git.circuitco.de/self/greyhouse/server"

	"git.circuitco.de/doras/authserver/interceptor"

	grpcm "github.com/grpc-ecosystem/go-grpc-middleware"
)

var bindAddr = "0.0.0.0:9999"

func main() {
	listen, err := net.Listen("tcp", bindAddr)
	if err != nil {
		log.Fatalf("Failed to listen to %s because %s", bindAddr, err)
	}

	log.Info("Starting service.")
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcm.ChainUnaryServer(interceptor.LogInterceptor))
	)
	nodeService := node.NewServer()
	api.RegisterNodeServer(server, nodeService)
	log.Info("Services listening forever.")
	server.Serve(listen)
	log.Fatal("Service is going down...")
}
