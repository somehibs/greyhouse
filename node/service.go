package node

import (
	"log"
	"time"
	"math/rand"
	"golang.org/x/net/context"

	api "git.circuitco.de/self/greyhouse/api"
)

type Node struct {
	Name string
	Address string
	Key string
	Room api.Room
	LastSeen time.Time
}

type NodeService struct {
	nodes map[string]*Node
}

var Instance = NodeService{}

func NewService() NodeService {
	return NodeService{nodes: map[string]*Node{}}
}

const keyChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
func randomKey(size int) string {
	key := make([]byte, size)
	for i := range key {
		key[i] = keyChars[rand.Int63()%int64(len(keyChars))]
	}
	return string(key)
}

func (ns NodeService) Register(ctx context.Context, metadata *api.NodeMetadata) (*api.NodeKey, error) {
	if ns.nodes[metadata.Identifier] != nil {
		// Already identified, check key. for now just return errors
		return &api.NodeKey{Key: ns.nodes[metadata.Identifier].Key}, nil
	}
	ns.nodes[metadata.Identifier] = &Node{
		Name: metadata.Identifier,
		Address: metadata.ClientAddress,
		Room: metadata.Room,
		LastSeen: time.Now(),
		Key: randomKey(25),
	}
	log.Printf("Register called: %+v\n", metadata)
	log.Printf("Stored: %+v\n", ns.nodes[metadata.Identifier])
	return &api.NodeKey{Key: ns.nodes[metadata.Identifier].Key}, nil
}
