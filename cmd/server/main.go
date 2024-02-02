package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const port = 8880

var templates = template.Must(template.ParseGlob("internal/templates/*.html"))

// routeMethod maps an http verb to a handler
type routeMethod map[string]http.HandlerFunc

// router is a map of paths to a collection of handlers identified by http verb
type router map[string]routeMethod

func renderTemplate(writer http.ResponseWriter, name string, data any) {
	err := templates.ExecuteTemplate(writer, name, data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func renderTemplateHandler(name string, data any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, name, data)
	}
}

func (r router) mapRoutes() {
	for route, handlers := range r {
		http.HandleFunc(route, func(writer http.ResponseWriter, request *http.Request) {
			handler, ok := handlers[request.Method]
			if !ok {
				http.Error(writer, "Not found", http.StatusNotFound)
				return
			}

			handler(writer, request)
		})
	}
}

var routes = router{
	"/": map[string]http.HandlerFunc{
		http.MethodGet: renderTemplateHandler("index.html", nil),
	},
}

func main() {
	addr := fmt.Sprintf(":%d", port)
	routes.mapRoutes()
	log.Fatal(http.ListenAndServe(addr, nil))
}
