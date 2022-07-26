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

func (v *View) Create(a *App, view View) {
	t := Template{
		base: a.Config.BaseTemplates,
	}
	// Add templates
	finalTemplates := t.Run(view.CurrentTemplates...)
	if len(finalTemplates) > 0 {
		log.Printf("%s: %s\n", view.CurrentRoute, finalTemplates)
	}
	// Add middleware
	var wrappedHandler http.HandlerFunc
	for _, m := range a.middleware {
		if m != nil {
			wrappedHandler = m(v.handleFuncWrapper(finalTemplates, a, view.CurrentView))
		}
	}

	// In case there is an option to turn off all gomek default middleware
	if wrappedHandler == nil {
		wrappedHandler = v.handleFuncWrapper(finalTemplates, a, view.CurrentView)
	}
	// Create handler
	a.mux.HandleFunc(view.CurrentRoute, wrappedHandler)
}

func stripTokens(pathSegment string) string {
	path := strings.Split(pathSegment, "<")
	path = strings.Split(path[1], ">")
	return path[0]
}

func getView(r *http.Request, a *App) (*View, map[string]string, bool) {
	for _, v := range a.view.StoredViews {
		if r.URL.Path == v.CurrentRoute {
			return &v, nil, true
		}
		// Check the path variables
		urlPaths := strings.Split(r.URL.Path, "/")[1:]
		// Check the root paths are the same
		if v.rootName == urlPaths[0] {
			// If this is just a single root/<arg> then return matched
			if len(urlPaths) == 2 {
				vars := map[string]string{
					stripTokens(v.routePaths[1]): urlPaths[1],
				}
				return &v, vars, true
			}
		}
	}
	return nil, nil, false
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

func (v *View) handleFuncWrapper(templates []string, a *App, currentView CurrentView) http.HandlerFunc {
	var data Data
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			routeMatches  bool
			methodMatches bool
			vars          map[string]string
		)
		// Route
		if v, viewVars, ok := getView(r, a); ok {
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
			te.ExecuteTemplate(w, a.Config.BaseTemplateName, data)
		} else {
			// No templates so treat as JSON / TEXT
			r.Header.Set("Content-Type", "application/json")
		}
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
			// If the are any path variables in a route then just break out and store the
			// first path segment as the root
			if string(r[i][0]) == "<" && string(r[i][len(r[i])-1]) == ">" {
				// Swap the caller's current route to registeredRoute
				c.registeredRoute = a.currentRoute
				// Register routes by storing route metadata in a View type
				// If a route has path variables, then set the route as the first path segment e.g /blogs/
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
