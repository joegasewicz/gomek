package gomek

import (
	"html/template"
	"log"
	"net/http"
)

type View struct {
	CurrentRoute     string
	CurrentMethods   []string
	CurrentTemplates []string
	CurrentView      CurrentView
	StoredViews      []View
}

func (v *View) Create(a *App, view View) {
	t := Template{
		base: a.baseTemplates,
	}
	finalTemplates := t.Run(view.CurrentTemplates...)
	a.mux.HandleFunc(view.CurrentRoute, v.handleFuncWrapper(finalTemplates, a, view.CurrentView))
}

func doesMethodMatch() {

}

func getView(r *http.Request, a *App) (*View, bool) {
	for _, v := range a.view.StoredViews {
		if r.URL.Path == v.CurrentRoute {
			return &v, true
		}

	}
	return nil, false
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
			view          View
		)
		// Route
		if v, ok := getView(r, a); ok {
			routeMatches = true
			// Methods
			if len(view.CurrentMethods) == 0 {
				// check route matches & skip method checks
				methodMatches = true
			}
			methodMatches = testMethod(r, *v)
		}
		if !routeMatches || !methodMatches {
			return
		}
		// Handler processes data only
		currentView(w, r, &data)
		// Add template(s)
		te, err := template.ParseFiles(templates...)
		if err != nil {
			log.Fatalf("Error parsing templates: %v", err.Error())
		}
		te.ExecuteTemplate(w, a.Config.BaseTemplateName, data)
	}
}

func (v *View) Store(a *App) {
	currView := View{
		CurrentRoute:     a.currentRoute,
		CurrentMethods:   a.currentMethods,
		CurrentTemplates: a.currentTemplates,
		CurrentView:      a.currentView,
	}
	v.StoredViews = append(v.StoredViews, currView)
}
