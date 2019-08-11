package node

import (
	"log"
	"time"
	"math/rand"
	"errors"
	"strings"
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/metadata"

	api "git.circuitco.de/self/greyhouse/api"
)

type Node struct {
	Name string
	Address string
	Key string
	Room api.Room
	LastSeen time.Time
	Modules []string
}

func (n Node) HasModule(_module string) bool {
	for _, module := range n.Modules {
		if _module == module {
			return true
		}
	}
	return false
}

type NodeService struct {
	Nodes map[string]*Node
}

func NewService() NodeService {
	return NodeService{Nodes: map[string]*Node{}}
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
	if ns.Nodes[metadata.Identifier] != nil {
		// Already identified, check key. for now just return errors
		return &api.NodeKey{Key: ns.Nodes[metadata.Identifier].Key}, nil
	}
	p, _ := peer.FromContext(ctx)
	clientAddress := strings.Split(p.Addr.String(), ":")[0]
	ns.Nodes[metadata.Identifier] = &Node{
		Name: metadata.Identifier,
		Address: clientAddress,
		Room: metadata.Room,
		LastSeen: time.Now(),
		Key: randomKey(25),
		Modules: metadata.Modules,
	}
	log.Printf("Register called: %+v\n", metadata)
	log.Printf("Stored: %+v\n", ns.Nodes[metadata.Identifier])
	return &api.NodeKey{Key: ns.Nodes[metadata.Identifier].Key}, nil
}

func AuthContext(ctx context.Context, key string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"node_key": key}))
}

func (ns NodeService) GetNode(ctx context.Context) *Node {
	data, _ := metadata.FromIncomingContext(ctx)
	return ns.getNodeInternal(data)
}

func (ns NodeService) getNodeInternal(metadata map[string][]string) *Node {
	for _, v := range ns.Nodes {
		if strings.Compare(v.Key, metadata["node_key"][0]) == 0 {
			return v
		}
	}
	return nil
}

func (ns NodeService) Check(addr string, metadata map[string][]string) error {
	if metadata["node_key"] == nil {
		return errors.New("authentication failed")
	}
	if ns.getNodeInternal(metadata) == nil {
		return errors.New("authentication failed")
	}
	return nil
}
