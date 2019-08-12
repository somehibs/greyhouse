package web

import (
	"net/http"
	"log"
	"html/template"

	api "git.circuitco.de/self/greyhouse/api"

	"git.circuitco.de/self/greyhouse/node"
)

var templates *template.Template

type HttpService struct {
	nodes *node.NodeService
}

var service HttpService

func Route(listenAddr string, nodeService *node.NodeService) {
	service = HttpService{nodeService}
	log.Print("Routing public-facing web assets.")
	// Handle some application routes.
	http.HandleFunc("/", service.webMain)
	http.HandleFunc("/cam", service.camMain)
	// Handle static resources.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	go http.ListenAndServe(listenAddr, nil)
}

type CamInfo struct {
	ImgSrc string
	Hlsl bool
	Name string
}

type CamView struct {
	Cameras []CamInfo
}

func (s *HttpService) camMain(w http.ResponseWriter, r *http.Request) {
	loadTemplates()
	view := CamView{}
	view.Cameras = make([]CamInfo, 0)
	for nodeName, node := range s.nodes.Nodes {
		if !node.HasModule("video") {
			log.Printf("Skipping node, missing video: %s", nodeName)
			continue
		}
		if nodeName == "bedroom" {
			//continue
		}
		view.Cameras = append(view.Cameras, CamInfo{"http://"+node.Address, false, nodeName})
	}
	view.Cameras = append(view.Cameras, CamInfo{"http://192.168.0.25", true, "printer"})
	log.Printf("Cameras: %+v", view)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	renderTemplateImpl(w, "cam", &view)
}

func renderTemplate(w http.ResponseWriter, tpl string) {
	renderTemplateImpl(w, tpl, nil)
}

func renderTemplateImpl(w http.ResponseWriter, tpl string, obj interface{}) {
	templates.ExecuteTemplate(w, "preamble", nil)
	templates.ExecuteTemplate(w, tpl, obj)
	templates.ExecuteTemplate(w, "postamble", nil)
}

func loadTemplates() {
	templates = template.Must(template.ParseFiles(
		"web/tpl/main",
		"web/tpl/cam",
		"web/tpl/preamble",
		"web/tpl/postamble",
		"web/tpl/light"))
}

type MainView struct {
	Rooms []RoomView
}

type RoomView struct {
	Name string
}

var renderRooms = []api.Room{api.Room_LOUNGE, api.Room_BEDROOM, api.Room_STUDY}

func (s *HttpService) webMain(w http.ResponseWriter, r *http.Request) {
	loadTemplates()
	view := MainView{make([]RoomView, 0)}
	for _, id := range renderRooms { // id, name
		room := RoomView{api.Room_name[int32(id)]}
		log.Printf("Room %+v", room)
		view.Rooms = append(view.Rooms, room)
	}
	renderTemplateImpl(w, "main", &view)
}
