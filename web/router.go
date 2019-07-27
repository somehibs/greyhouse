package web

import (
	"net/http"
	"log"
	"html/template"
)

var templates *template.Template

func Route(listenAddr string) {
	log.Print("Routing public-facing web assets.")
	// Handle some application routes.
	http.HandleFunc("/", webMain)
	// Handle static resources.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	go http.ListenAndServe(listenAddr, nil)
}

func webMain(w http.ResponseWriter, r *http.Request) {
	templates = template.Must(template.ParseFiles(
		"web/tpl/main",
		"web/tpl/preamble",
		"web/tpl/postamble",
		"web/tpl/light"))
	templates.ExecuteTemplate(w, "preamble", nil)
	templates.ExecuteTemplate(w, "main", nil)
	templates.ExecuteTemplate(w, "postamble", nil)
}
