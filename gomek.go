package gomek

import (
	"context"
	"errors"
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

// Data type used to reference the data reference from the handler args
//
//	func MyHandler(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
//		*d = map[string]string{ "v": "k"}
type Data map[string]interface{}

// CurrentView is a custom type representing the http.HandlerFunc type.
type CurrentView func(http.ResponseWriter, *http.Request, *Data)

type Handle func(pattern string, handler http.Handler)

type Middleware []func(http.Handler) http.HandlerFunc

// Config type that should be passed to `gomek.New`
type Config struct {
	BaseTemplateName string
	BaseTemplates    []string
}

type Resource interface {
	Delete(http.ResponseWriter, *http.Request, *Data)
	Get(http.ResponseWriter, *http.Request, *Data)
	Post(http.ResponseWriter, *http.Request, *Data)
	Put(http.ResponseWriter, *http.Request, *Data)
}

type IApp interface {
	resetCurrentView()
	cloneRoute()
	Start() error
	SetHost(host string)
	Listen(port int)
	Methods(methods ...string) *App
	Route(route string) *App
	Templates(templates ...string)
	BaseTemplates(templates ...string)
	View(view CurrentView) *App
	Resource(m Resource) *App
	Use(h func(http.Handler) http.HandlerFunc)
	Shutdown()
	GetView() *View
	GetConfig() *Config
}

type App struct {
	baseTemplateName string
	baseTemplates    []string
	Config           Config
	currentRoute     string
	currentMethods   []string
	currentTemplates []string
	currentView      CurrentView
	currentResource  Resource
	mux              *http.ServeMux
	Host             string
	Port             int
	Protocol         string
	view             View
	Handle           Handle
	middleware       Middleware
	rootCtx          context.Context
	authCtx          context.Context
	server           *http.Server
}

func createAddr(a *App) string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

func (a *App) GetView() *View {
	return &a.view
}

func (a *App) GetConfig() *Config {
	return &a.Config
}

func (a *App) setup() *http.Server {
	var auth = map[string]string{}
	// Set app context
	a.rootCtx = context.Background()
	a.authCtx = context.WithValue(a.rootCtx, "auth", auth)
	// Store the last registered view
	if a.currentResource != nil {
		a.view.StoreResource(a)
		// Duplicate the resource methods for /<path_name>/ to /<path_name>
		// This is because Go's http package swaps out POSTs to GETS with a /<path_name>/ path.
		a.cloneRoute()
	} else {
		a.view.Store(a)
	}
	if len(a.currentMethods) < 1 {
		a.currentMethods = DEFAULT_METHODS
	}
	a.resetCurrentView()
	// Handle defaults
	if a.Config.BaseTemplateName == "" {
		a.Config.BaseTemplateName = DEFAULT_BASE_TEMPLATE
	}
	if a.Host == "" {
		a.Host = DEFAULT_HOST
	}
	if a.Port == 0 {
		a.Port = DEFAULT_PORT
	}
	if a.Protocol == "" {
		a.Protocol = DEFAULT_PROTOCOL
	}
	// Create views
	for _, v := range a.view.StoredViews {
		a.view.Create(&a.Config, &a.middleware, a.mux, v)
	}

	// Create the origin
	address := createAddr(a)
	// Server
	return &http.Server{
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
}

func (a *App) resetCurrentView() {
	a.currentRoute = ""
	a.currentMethods = nil
	a.baseTemplates = nil
	a.currentView = nil
}

func (a *App) cloneRoute() {
	// Duplicate the resource methods for /<path_name>/ to /<path_name>
	// This is because Go's http package swaps out POSTs to GETS with a /<path_name>/ path.
	if a.currentRoute[len(a.currentRoute)-1:] == ">" {
		for _, m := range a.currentMethods {
			if m == "POST" {
				// Construct a path name from the stored `registeredRoute` value
				for _, storedView := range a.view.StoredViews {
					if storedView.registeredRoute == a.currentRoute {
						// Create a route without the slash at the parth end e.g /<path_name>
						a.currentRoute = fmt.Sprintf("/%s", storedView.rootName)
						a.view.StoreResource(a)
					}
				}
				break
			}
		}
	}
}

// SetHost sets the host. Default is ":"
// If no port is set, then Gomek will default to `5000`
//
//	app.SetHost("localhost")
func (a *App) SetHost(host string) {
	a.Host = host
}

// Listen sets the port the server will accept request on.
// If no port is set, then Gomek will default to `5000`
//
//	app.Listen(5001)
func (a *App) Listen(port int) {
	a.Port = port
}

// Methods CRUD methods to match on the request URL. If there are no methods
// declared, then it defaults to - `"GET"`
// For Example
//
//	app.Methods("GET")
//	app.Methods("GET", "POST", "DELETE")
func (a *App) Methods(methods ...string) *App {
	if len(methods) == 0 {
		methods = append(methods, "GET")
	}
	// Add OPTIONS
	methods = append(methods, "OPTIONS")
	a.currentMethods = methods
	return a
}

// Route A string representing the incoming request URL.
// This is the first argument to Gomek's mux.Route() method or the first
// argument to http.HandleFunc(). For Example
//
//	app.Route("/") // ... other chained methods
func (a *App) Route(route string) *App {
	if a.currentRoute != "" {
		// Store the previous view to lazily register them at run time
		if a.currentResource != nil {
			a.view.StoreResource(a)
			//a.cloneRoute()
		} else {
			a.view.Store(a)
		}
	}
	// This route gets registered in the Start method
	a.currentRoute = route
	return a
}

// Templates Method that takes a single template relative path or multiple template
// path slices of main route templates (not partial templates). For example:
//
//	 app.Route("/")
//		.View(Home)
//		.Methods("GET")
//		.Templates("./templates/hero.html", "./templates/routes/home.html")
//
// The above example adds a `hero.html` partial template & a main route `home.html` template.
func (a *App) Templates(templates ...string) {
	a.currentTemplates = templates
}

// BaseTemplates method accepts slices of string, string if the name of the
// template file.
//
//	baseTemplates := []string{
//		"./templates/layout.html",
//		"./templates/sidebar.html",
//		"./templates/navbar.html",
//		"./templates/footer.html",
//	}
//	app.BaseTemplate(baseTemplates)
func (a *App) BaseTemplates(templates ...string) {
	a.Config.BaseTemplates = templates
}

// View is called if the Route request URL is matched.
// handler arg is your View function.Create a View - template data needs to be passed
// by value to `data *map[string]interface{}`
//
//			func Home(w http.ResponseWriter, r *http.Request, data *gomek.Data) {
//	   		var templateData gomek.Data // Or map[string]interface{} .etc... // Create a map to store your template data
//	   		templateData = make(map[string]interface{})
//	   		templateData["heading"] = "Create a new advert"
//	   		*data = templateData // pass by value back to `data`
//			}
//
//			app.
//			  // ...
//			  .View(Home)
//			  // ...
func (a *App) View(view CurrentView) *App {
	a.currentView = view
	return a
}

// Resource accepts a type that implements the `Resource` interface.
// This is useful for Rest design handler methods attached to a named resource
// type.
//
//	type Notice struct {
//	}
//
//	func (n *Notice) Post(w http.ResponseWriter, request *http.Request, data *gomek.Data) {
//		panic("implement me")
//	}
//
//	func (n *Notice) Put(w http.ResponseWriter, request *http.Request, data *gomek.Data) {
//		panic("implement me")
//	}
//
//	func (n *Notice) Delete(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
//		panic("implement me")
//	}
//
//	func (n *Notice) Get(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
//		var notice schemas.Notice
//		notice.Name = "Joe!"
//		gomek.JSON(w, notice, http.StatusOK)
//	}
//
// To use your implentation of the `Resource` type
//
//	app.Route("/notices").Resource(&routes.Notice{}).Methods("GET")
func (a *App) Resource(m Resource) *App {
	a.currentResource = m
	return a
}

// Use adds middleware.
//
//	app := gomek.New(gomek.Config{})
//	app.Use(gomek.CORS)
func (a *App) Use(h func(http.Handler) http.HandlerFunc) {
	a.middleware = append(a.middleware, h)
}

// Shutdown force shutdown of the mux server
//
//	app.Shutdown()
func (a *App) Shutdown() {
	err := a.server.Shutdown(a.rootCtx)
	if err != nil {
		log.Fatalln("error shutting down", err)
	}
}

// Args access the request arguments in a handler as a map
//
//	args := gomek.Args(r)
func Args(r *http.Request) map[string]string {
	if vars := r.Context().Value("uriArgs"); vars != nil {
		return vars.(map[string]string)
	}
	return nil
}

// GetParams returns slices of string
//
//	// example request url - http://127.0.0.1:8080/users?user_id=1
//	userID, err := gomek.QueryParams(r, "user_id")
//	if err != nil {
//		log.Println("no user_id in params")
//		return
//	}
//	// userID[0] = "1"
func GetParams(r *http.Request, name string) ([]string, error) {
	params := r.URL.Query()
	paramValue, present := params[name]
	if !present || len(paramValue) == 0 {
		log.Println("no noticeboardID in params")
		return nil, errors.New("param not" + name + " present")
	}
	return paramValue, nil
}

// App
type _App struct {
	*App
}

// Start sets up all the registered views, templates & middleware
//
//	app = gomek.New(gomek.Config{})
//	app.Start()
func (a *App) Start() error {
	// Start server...
	a.server = a.setup()
	log.Printf("Starting server on %s://%s", a.Protocol, a.server.Addr)
	err := a.server.ListenAndServe()
	if err != nil {
		log.Println("error starting gomek server", err)
	} else {
		log.Printf("Starting server on %s://%s", a.Protocol, a.server.Addr)
	}
	return err
}

// New creates a new gomek application
//
//	app := gomek.New(gomek.Config{})
func New(config Config) *App {
	mux := http.NewServeMux()
	return &App{
		Config: config,
		mux:    mux,
		Handle: mux.Handle,
	}
}

type TestApp struct {
	App
}

// NewTestApp creates a new gomek application
//
//	app := gomek.NewTestApp(gomek.Config{})
func NewTestApp(config Config) IApp {
	mux := http.NewServeMux()
	app := TestApp{
		App{
			Config: config,
			mux:    mux,
			Handle: mux.Handle,
		},
	}
	return &app
}

func (a *TestApp) Start() error {
	a.App.setup()
	return nil
}
