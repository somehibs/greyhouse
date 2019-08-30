package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	api "git.circuitco.de/self/greyhouse/api"
	"git.circuitco.de/self/greyhouse/version"

	"git.circuitco.de/self/greyhouse/modules"
)

type ClientConfig struct {
	Node string
	NodeAddress string
	Room api.Room
	Server string
	Modules []modules.ModuleConfig
}

func loadClientConfig() (ClientConfig, error) {
	clientConfig := ClientConfig{}
	f, err := os.Open("client.json")
	if err != nil {
		return clientConfig, err
	}
	read, err := ioutil.ReadAll(f)
	if err != nil {
		return clientConfig, err
	}
	err = json.Unmarshal(read, &clientConfig)
	if clientConfig.Room == 0 {
		return clientConfig, errors.New("room not correctly set")
	}
	return clientConfig, err

}

var bindAddr = "0.0.0.0:9991" // not implemented
var thisVersion = version.CurrentVersion()
var loadedModules = make([]modules.GreyhouseClientModule, 0)
var tickModules = make([]modules.GreyhouseClientModule, 0)
var moduleNames = make([]string, 0)

func loadModules(moduleConfig []modules.ModuleConfig) error {
	log.Print("loading modules")
	var err error
	loadedModules, err = modules.LoadModules(moduleConfig)
	if err != nil {
		return err
	}
	moduleNames = make([]string, 0)
	for _, cfg := range moduleConfig {
		moduleNames = append(moduleNames, cfg.Name)
	}
	for _, module := range loadedModules {
		if module.CanTick() {
			tickModules = append(tickModules, module)
		}
	}
	shutdownSignal()
	log.Print("loaded")
	return nil
}

func shutdownSignal() {
	// trap signals
	log.Print("trapping shutdown to allow for module shutdown...")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		for _, module := range loadedModules {
			module.Shutdown()
		}
		panic("Interrupted.")
	}()
}

func registered(clientHost modules.ClientHost) {
	log.Print("Connected. Ticking modules...")
	// refresh the modules
	modules.SetClientHost(&clientHost)
	for _, module := range loadedModules {
		module.Update()
	}
	// spin on ticking unless an error comes back about networking
	tickCount := 0
	for ;; {
		for _, module := range tickModules {
			e := module.Tick()
			if e != nil {
				log.Printf("FATAL Could not tick module due to %s", e.Error())
			}
		}
		time.Sleep(1*time.Second)
		tickCount += 1
		if tickCount % 5 == 0 {
			for _, module := range loadedModules {
				e := module.Tick()
				if e != nil {
					log.Printf("all modules tick found %s in module", e.Error())
					return
				}
			}
		}
	}
}

func getClients(conn *grpc.ClientConn, nodeClient *api.PrimaryNodeClient, nodeKey string) modules.ClientHost {
	ch := modules.ClientHost{Node: nodeClient, Key: nodeKey}
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
	log.Print("started")
	log.Print("loading config...")
	config, err := loadClientConfig()
	if err != nil {
		panic("Could not load config json: " + err.Error())
	}
	err = loadModules(config.Modules)
	if err != nil {
		log.Panicf("Failed to load a module and ending safely now %+v", err)
		return
	}

	for ;; {
		conn, err := grpc.Dial(config.Server, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect to %s: %s", config.Server, err.Error())
		}

		nodeClient := api.NewPrimaryNodeClient(conn)
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		//for ;; {
		//	log.Fatal("need some modules to process")
		//}

		// Call Register with our favourite address
		i, e := nodeClient.Register(ctx, &api.NodeMetadata{
			Identifier: config.Node,
			ClientAddress: config.NodeAddress,
			Room: config.Room,
			Version: &thisVersion,
			Modules: moduleNames,
		})
		if e == nil {
			// Perfect, we connected ok
			rc := api.NewRulesClient(conn)
			clientHost := getClients(conn, &nodeClient, i.Key)
			l, err := rc.List(clientHost.GetContext(), &api.RuleFilter{})
			if err != nil {
				log.Print("Could not get list: " + err.Error())
			} else {
				log.Printf("Rule list: %+v\n", l)
			}
			registered(clientHost)
			log.Print("Warning: registered() returned, authentication failure? retry connection")
		} else {
			log.Printf("R: %+v Error: %+v\n", i, e)
		}
		time.Sleep(4*time.Second)
	}
}

