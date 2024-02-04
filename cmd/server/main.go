package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
)

const port = 8880

// routeMethod maps an http verb to a handler
type routeMethod map[string]http.HandlerFunc

// router is a map of paths to a collection of handlers identified by http verb
type router map[string]routeMethod

func templatesWithLayout(names ...string) *template.Template {
	getPath := func(filename string) string { return path.Join("internal", "templates", filename) }

	filenames := make([]string, len(names)+1)
	filenames[0] = getPath("layout.gohtml")
	for i, name := range names {
		filenames[i+1] = getPath(fmt.Sprintf("%s.gohtml", name))
	}

	return template.Must(template.ParseFiles(filenames...))
}

func renderTemplates(writer http.ResponseWriter, data any, t *template.Template) {
	err := t.ExecuteTemplate(writer, "layout", data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func renderTemplatesHandler(data any, names ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := templatesWithLayout(names...)
		renderTemplates(w, data, t)
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
		http.MethodGet: renderTemplatesHandler(nil, "home"),
	},
}

func main() {
	addr := fmt.Sprintf(":%d", port)
	routes.mapRoutes()
	log.Fatal(http.ListenAndServe(addr, nil))
}
