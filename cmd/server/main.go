package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path"
)

const port = 8880

// router is a map of routes to a collection of handlers identified by http verb
type router map[string]http.HandlerFunc

type page struct {
	Name string
}

func templatesWithLayout(names ...string) *template.Template {
	getPath := func(filename string) string { return path.Join("ui", "templates", filename) }

	filenames := make([]string, len(names)+2)
	filenames[0] = getPath("layout.gohtml")
	filenames[1] = getPath("nav.gohtml")
	for i, name := range names {
		filenames[i+2] = getPath(fmt.Sprintf("%s.gohtml", name))
	}

	return template.Must(template.ParseFiles(filenames...))
}

func renderTemplates(writer http.ResponseWriter, data any, t *template.Template) {
	err := t.ExecuteTemplate(writer, "layout", data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func renderTemplatesHandler(data page, names ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := templatesWithLayout(names...)
		renderTemplates(w, data, t)
	}
}

func (r router) mapRoutes() {
	for route, handler := range r {
		http.HandleFunc(route, handler)
	}
}

func routeStaticFiles() {
	root, _ := url.Parse("/static/")
	cssPath := root.JoinPath("css/").Path
	imgPath := root.JoinPath("images/").Path
	http.Handle(get(cssPath), http.StripPrefix(cssPath, http.FileServer(http.Dir("ui/static/css"))))
	http.Handle(get(imgPath), http.StripPrefix(imgPath, http.FileServer(http.Dir("ui/static/images"))))
}

func routePattern(method string, path string) string {
	return fmt.Sprintf("%s %s", method, path)
}

func get(path string) string {
	return routePattern(http.MethodGet, path)
}

var routes = router{
	get("/{$}"):     renderTemplatesHandler(page{Name: "Home"}, "home"),
	get("/contact"): renderTemplatesHandler(page{Name: "Contact"}, "contact"),
}

func main() {
	routeStaticFiles()
	routes.mapRoutes()

	addr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}
