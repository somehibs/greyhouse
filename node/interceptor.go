package node

import (
	"strings"
	"errors"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	// "google.golang.org/grpc/peer"
	// "google.golang.org/grpc/status"
)

var AllowedMethods = map[string]bool{}

func NodeInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if AllowedMethods[info.FullMethod] == true {
		return handler(ctx, req)
	}
	data, _ := metadata.FromIncomingContext(ctx)
	if data["node_key"] == nil {
		return nil, errors.New("authentication failed")
	}
	ok := false
	for k := range Instance.nodes {
		if strings.Compare(Instance.nodes[k].Key, data["node_key"][0]) == 0 {
			ok = true
		}
	}
	if ok == false {
		return nil, errors.New("authentication failed")
	}
	return handler(ctx, req)
}
