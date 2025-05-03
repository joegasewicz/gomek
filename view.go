package gomek

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type ViewTemplate struct {
	Route     string
	Templates []string
}

type View struct {
	registeredRoute string
	routePaths      []string
	rootName        string
	Route           string
	Methods         []string
	Templates       []string
	View            CurrentView
	StoredViews     []View
}

// NewTestView view testing util. Enables the testing of individual

func NewTestView(views []View) View {
	return View{
		registeredRoute: "",
		routePaths:      nil,
		rootName:        "",
		Route:           "/blogs",
		Methods:         []string{"POST"},
		Templates:       nil,
		View:            nil,
		StoredViews:     views,
	}
}

func (v *View) Create(a *App, view View) {
	var finalTemplates []string
	// Validate view
	if view.Route == "" {
		log.Println("[GOMEK] Warning: Route is set to an empty string!")
	}

	t := Template{
		base: a.Config.BaseTemplates,
	}
	// Add registeredTemplates
	if len(view.Templates) > 0 {
		finalTemplates = t.Run(view.Templates...)
		// Add the route & templates for logging
		registerTemplate := RegisteredTemplates{
			Route:     view.Route,
			Templates: view.Templates,
			Partials:  t.base,
		}
		a.registeredTemplates = append(a.registeredTemplates, registerTemplate)
	}

	// Add middleware
	var wrappedHandler http.HandlerFunc
	wrappedHandler = v.handleFuncWrapper(finalTemplates, &a.Config, view, view.View)

	for _, m := range a.middleware {
		if m != nil {
			wrappedHandler = m(wrappedHandler)
		}
	}

	// In case there is an option to turn off all gomek default middleware
	if wrappedHandler == nil {
		wrappedHandler = v.handleFuncWrapper(finalTemplates, &a.Config, view, view.View)
	}
	// Create handler
	a.Mux.HandleFunc(view.Route, wrappedHandler)
}

func stripTokens(pathSegment string) string {
	path := strings.Split(pathSegment, "<")
	path = strings.Split(path[1], ">")
	return path[0]
}

func parseView(r *http.Request, view View) (*View, map[string]string, bool) {
	if r.URL.Path == view.Route {
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
	if view.View != nil {
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
	for _, method := range v.Methods {
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
			if len(v.Methods) == 0 {
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
				out := fmt.Sprintf("[GOMEK]: Error parsing registeredTemplates: %v", err.Error())
				out = PrintWithColor(out, RED)
				log.Fatalf(out)
			}
			err = te.ExecuteTemplate(w, config.BaseTemplateName, data)
			if err != nil {
				log.Printf("[GOMEK] Error: Error executing template!\n %e", err)
			}
		} else {
			// No registeredTemplates so treat as JSON / TEXT
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
		Route:     a.currentRoute,
		Methods:   a.currentMethods,
		Templates: a.currentTemplates,
		View:      a.currentView,
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
				c.Route = fmt.Sprintf("/%s/", r[1])
				// Save the path segments in slices of string, then we can match the path variables against incoming request
				c.routePaths = r[1:]
				c.rootName = r[1]
				break
			}
		}
	}

	v.StoredViews = append(v.StoredViews, c)
	a.resetCurrentView()
}
