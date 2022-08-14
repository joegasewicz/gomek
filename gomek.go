package gomek

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

const (
	DEFAULT_BASE_TEMPLATE = "layout"
	DEFAULT_HOST          = "localhost"
	DEFAULT_PORT          = 5000
	DEFAULT_PROTOCOL      = "http"
)

var (
	DEFAULT_METHODS = []string{"GET"}
)

type Gomek interface {
	Start()
	Listen(port int)
	Methods(methods ...string) *App
	Route(route string) *App
	Templates(templates ...string)
	BaseTemplates(templates ...string)
	New(config Config) App
	View(view CurrentView) *App
	Use(h func(http.Handler) http.HandlerFunc)
	Args(r *http.Request) map[string]string
}

// Data type used to reference the data reference from the handler args
//
//
//		func MyHandler(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
//			*d = map[string]string{ "v": "k"}
//
type Data map[string]interface{}

// CurrentView is a custom type representing the http.HandlerFunc type.
type CurrentView func(http.ResponseWriter, *http.Request, *Data)

type Handle func(pattern string, handler http.Handler)

// Config type that should be passed to `gomek.New`
type Config struct {
	BaseTemplateName string
	BaseTemplates    []string
}

// App
type App struct {
	baseTemplateName string
	baseTemplates    []string
	Config           Config
	currentRoute     string
	CurrentMethods   []string
	currentTemplates []string
	currentView      CurrentView
	mux              *http.ServeMux
	CurrentHost      string
	Port             int
	Protocol         string
	view             View
	Handle           Handle
	middleware       []func(http.Handler) http.HandlerFunc
	rootCtx          context.Context
	authCtx          context.Context
	server           *http.Server
}

func createAddr(a *App) string {
	return fmt.Sprintf("%s:%d", a.CurrentHost, a.Port)
}

func (a *App) resetCurrentView() {
	a.currentRoute = ""
	a.CurrentMethods = nil
	a.baseTemplates = nil
	a.currentView = nil
}

// Start sets up all the registered views, templates & middleware
//
//
//		app = gomek.New(gomek.Config{})
//		app.Start()
//
func (a *App) Start() error {
	var auth = map[string]string{}
	// Set app context
	a.rootCtx = context.Background()
	a.authCtx = context.WithValue(a.rootCtx, "auth", auth)
	// Store the last registered view
	a.view.Store(a)
	a.resetCurrentView()
	// Handle defaults
	if a.Config.BaseTemplateName == "" {
		a.Config.BaseTemplateName = DEFAULT_BASE_TEMPLATE
	}
	if a.CurrentHost == "" {
		a.CurrentHost = DEFAULT_HOST
	}
	if a.Port == 0 {
		a.Port = DEFAULT_PORT
	}
	if a.Protocol == "" {
		a.Protocol = DEFAULT_PROTOCOL
	}
	if len(a.CurrentMethods) < 1 {
		a.CurrentMethods = DEFAULT_METHODS
	}
	// Add default middleware
	//a.Use(Logging)
	// Create views
	for _, v := range a.view.StoredViews {
		a.view.Create(a, v)
	}
	// Create the origin
	address := createAddr(a)
	// Server
	server := &http.Server{
		Addr:              address,
		Handler:           a.mux,
		TLSConfig:         nil,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}
	a.server = server
	// Start server...
	err := server.ListenAndServe()
	if err != nil {
		log.Println("error starting gomek server", err)
	} else {
		log.Printf("Starting server on %s://%s", a.Protocol, address)
	}
	return err
}

// Listen sets the port the server will accept request on.
// If no port is set, then Gomek will default to `5000`
//
//
// 		app.Listen(5001)
//
func (a *App) Listen(port int) {
	a.Port = port
}

// Host sets the CurrentHost
//
//
//		app.Host("localhost")
//
func (a *App) Host(h string) {
	a.CurrentHost = h
}

// Methods CRUD methods to match on the request URL. If there are no methods
// declared, then it defaults to - `"GET"`
// For Example
//
// 		app.Methods("GET")
//		app.Methods("GET", "POST", "DELETE")
//
func (a *App) Methods(methods ...string) *App {
	if len(methods) == 0 {
		methods = append(methods, "GET")
	}
	// Add OPTIONS
	methods = append(methods, "OPTIONS")
	a.CurrentMethods = methods
	return a
}

// Route A string representing the incoming request URL.
// This is the first argument to Gomek's mux.Route() method or the first
// argument to http.HandleFunc(). For Example:
//
//
//		app.Route("/") // ... other chained methods
//
func (a *App) Route(route string) *App {
	if a.currentRoute != "" {
		// Store the previous view to lazily register them at run time
		a.view.Store(a)
	}
	// This route gets registered in the Start method
	a.currentRoute = route
	return a
}

// Templates Method that takes a single template relative path or multiple template
// path slices of main route templates (not partial templates). For example:
//
//		 app.Route("/")
//			.View(Home)
//			.Methods("GET")
//			.Templates("./templates/hero.html", "./templates/routes/home.html")
//
// The above example adds a `hero.html` partial template & a main route `home.html` template.
func (a *App) Templates(templates ...string) {
	a.currentTemplates = templates
}

// BaseTemplates method accepts slices of string, string if the name of the
// template file.
//
//
//		baseTemplates := []string{
//			"./templates/layout.html",
//			"./templates/sidebar.html",
//			"./templates/navbar.html",
//			"./templates/footer.html",
//		}
//		app.BaseTemplate(baseTemplates)
//
func (a *App) BaseTemplates(templates ...string) {
	a.Config.BaseTemplates = templates
}

// New creates a new gomek application
//
//
//		app := gomek.New(gomek.Config{})
//
func New(config Config) App {
	mux := http.NewServeMux()

	return App{
		Config: config,
		mux:    mux,
		Handle: mux.Handle,
	}
}

// View is called if the Route request URL is matched.
// handler arg is your View function.Create a View - template data needs to be passed
// by value to `data *map[string]interface{}`
//
//
//		func Home(w http.ResponseWriter, r *http.Request, data *gomek.Data) {
//    		var templateData gomek.Data // Or map[string]interface{} .etc... // Create a map to store your template data
//    		templateData = make(map[string]interface{})
//    		templateData["heading"] = "Create a new advert"
//    		*data = templateData // pass by value back to `data`
//		}
//
//		app.
// 		  // ...
//		  .View(Home)
//		  // ...
//
func (a *App) View(view CurrentView) *App {
	a.currentView = view
	return a
}

// Use adds middleware.
//
//
//		app := gomek.New(gomek.Config{})
//		app.Use(gomek.CORS)
//
func (a *App) Use(h func(http.Handler) http.HandlerFunc) {
	a.middleware = append(a.middleware, h)
}

// Args access the request arguments in a handler as a map
//
//
// 		args := gomek.Args(r)
//
func Args(r *http.Request) map[string]string {
	if vars := r.Context().Value("uriArgs"); vars != nil {
		return vars.(map[string]string)
	}
	return nil
}

func (a *App) Shutdown() {
	err := a.server.Shutdown(a.rootCtx)
	if err != nil {
		log.Fatalln("error shutting down", err)
	}
}
