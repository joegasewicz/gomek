package gomek

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type View struct {
	registeredRoute  string
	routePaths       []string
	rootName         string
	CurrentRoute     string
	CurrentMethods   []string
	CurrentTemplates []string
	CurrentView      CurrentView
	StoredViews      []View
}

// NewTestView view testing util. Enables the testing of individual

func NewTestView(views []View) View {
	return View{
		registeredRoute:  "",
		routePaths:       nil,
		rootName:         "",
		CurrentRoute:     "/blogs",
		CurrentMethods:   []string{"POST"},
		CurrentTemplates: nil,
		CurrentView:      nil,
		StoredViews:      views,
	}
}

func (v *View) Create(config *Config, middleware *Middleware, mux *http.ServeMux, view View) {
	// Validate view
	if view.CurrentRoute == "" {
		log.Println("Route is not set")
		return
	}

	t := Template{
		base: config.BaseTemplates,
	}
	// Add templates
	finalTemplates := t.Run(view.CurrentTemplates...)
	if len(finalTemplates) > 0 {
		log.Printf("Registering templates: %s: %s\n", view.CurrentRoute, finalTemplates)
	}
	// Add middleware
	var wrappedHandler http.HandlerFunc
	wrappedHandler = v.handleFuncWrapper(finalTemplates, config, view, view.CurrentView)

	for _, m := range *middleware {
		if m != nil {
			wrappedHandler = m(wrappedHandler)
		}
	}

	// In case there is an option to turn off all gomek default middleware
	if wrappedHandler == nil {
		wrappedHandler = v.handleFuncWrapper(finalTemplates, config, view, view.CurrentView)
	}
	// Create handler
	mux.HandleFunc(view.CurrentRoute, wrappedHandler)
}

func stripTokens(pathSegment string) string {
	path := strings.Split(pathSegment, "<")
	path = strings.Split(path[1], ">")
	return path[0]
}

func parseView(r *http.Request, view View) (*View, map[string]string, bool) {
	if r.URL.Path == view.CurrentRoute {
		return &view, nil, true
	}
	// Check the path variables
	urlPaths := strings.Split(r.URL.Path, "/")[1:]
	// Check the root paths are the same
	if view.rootName == urlPaths[0] {
		// If this is just a single root/<arg> then return matched
		if len(urlPaths) == 2 {
			vars := map[string]string{
				stripTokens(view.routePaths[1]): urlPaths[1],
			}
			return &view, vars, true
		}
	}
	return nil, nil, false
}

func getView(r *http.Request, view View) (vv *View, mm map[string]string, b bool) {
	if view.CurrentView != nil {
		vv, mm, b = parseView(r, view)
	}

	for _, v := range view.StoredViews {
		vv, mm, b = parseView(r, v)
	}
	return vv, mm, b
}

func setViewVars(r *http.Request, vars map[string]string) *http.Request {
	ctx := context.WithValue(r.Context(), "uriArgs", vars)
	return r.WithContext(ctx)
}

func testMethod(r *http.Request, v View) bool {
	for _, method := range v.CurrentMethods {
		if method == r.Method {
			return true
		}
	}
	return false
}

func (v *View) handleFuncWrapper(templates []string, config *Config, view View, currentView CurrentView) http.HandlerFunc {
	var data Data
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			routeMatches  bool
			methodMatches bool
			vars          map[string]string
		)
		// Route
		if v, viewVars, ok := getView(r, view); ok {
			vars = viewVars
			routeMatches = true
			// Methods
			if len(v.CurrentMethods) == 0 {
				// check route matches & skip method checks
				methodMatches = true
			}
			methodMatches = testMethod(r, *v)
		}
		if !routeMatches || !methodMatches {
			return
		}
		// set context
		r = setViewVars(r, vars)
		// Handler processes data only
		currentView(w, r, &data)
		// Add template(s) if they exist
		if len(templates) > 0 {
			te, err := template.ParseFiles(templates...)
			if err != nil {
				log.Fatalf("Error parsing templates: %v", err.Error())
			}
			te.ExecuteTemplate(w, config.BaseTemplateName, data)
		} else {
			// No templates so treat as JSON / TEXT
			r.Header.Set("Content-Type", "application/json")
		}
	}
}

func (v *View) createHandlerFromResource(delete, get, post, put CurrentView) CurrentView {
	// Creates a single handler with a switch to call each resource declared function
	return func(w http.ResponseWriter, r *http.Request, d *Data) {
		switch r.Method {
		case "DELETE":
			{
				delete(w, r, d)
			}
		case "GET":
			{
				get(w, r, d)
			}
		case "POST":
			{
				post(w, r, d)
			}
		case "PUT":
			{
				put(w, r, d)
			}
		}
	}
}

func (v *View) StoreResource(a *App) {
	// Pull off each method string from the resource
	if a.currentResource != nil {
		var methods []string
		for _, m := range a.currentMethods {
			methods = append(methods, m)
		}
		// Build a handler that conditionally evokes the resource methods
		a.currentView = v.createHandlerFromResource(
			a.currentResource.Delete,
			a.currentResource.Get,
			a.currentResource.Post,
			a.currentResource.Put,
		)
		a.currentMethods = methods
		v.Store(a)
	}
}

func (v *View) Store(a *App) {
	c := View{
		CurrentRoute:     a.currentRoute,
		CurrentMethods:   a.currentMethods,
		CurrentTemplates: a.currentTemplates,
		CurrentView:      a.currentView,
	}
	if a.currentRoute != "/" {
		r := strings.Split(a.currentRoute, "/")
		for i := 1; i < len(r); i++ {
			// If there are any path variables in a route then just break out and store the
			// first path segment as the root
			if string(r[i][0]) == "<" && string(r[i][len(r[i])-1]) == ">" {
				// Swap the caller's current route to registeredRoute
				c.registeredRoute = a.currentRoute
				// Register routes by storing route metadata in a View type
				// If a route has path variables, then set the route as the first path segment e.g /blogs
				c.CurrentRoute = fmt.Sprintf("/%s/", r[1])
				// Save the path segments in slices of string, then we can match the path variables against incoming request
				c.routePaths = r[1:]
				c.rootName = r[1]
				break
			}
		}
	}

	v.StoredViews = append(v.StoredViews, c)
}
