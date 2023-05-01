# Gomek
A minimal http framework that includes some useful tools. Inspired by Flask.

⚠️ *Production ready in v1.0.0*

## Features
- Easy to learn API.
- Views handle your data only e.g. no template logic (Single-responsibility principle).
- Define HTTP methods per View method.
- Middleware - add yor own middleware
- Auth
- Logging
- CORS
- JSON
- Access request arguments
- Static files

# Install
```bash
go get github.com/joegasewicz/gomek
```

# Example
```go
c := gomek.Config{
    BaseTemplateName: "layout",
    BaseTemplates: []string{
        "./templates/layout.gohtml",
    },
}
// Create a new gomek app
app := gomek.New(c)
// Declare your views
app.Route("/blog").View(blog).Methods("GET", "POST").Templates("./templates/blog.gohtml")
// Add middleware
app.Use(gomek.Logging) // Use gomek's Logging
app.Use(gomek.CORS) // Use gomek's CORS
// Set a port (optional)
app.Listen(6011)
// Start your app
app.Start()
```

### Handlers
There are 2 types of handlers
- Handlers that render templates
- Handlers that return JSON / text .etc

```go
// A view that handles a template's data
func index(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
    templateData := make(gomek.Data)
	// templateData["Title"] = "Hello!"
	*d = templateData
}

// A view that just returns JSON
func blog(w http.ResponseWriter, r *http.Request, data *gomek.Data) {
    var blog Blog
	// query database then return JSON
	gomek.JSON(w, blog, http.StatusOK)
}
```

### Request Arguments
Access the request route arguments inside a handler
```go
args := gomek.Args(r)
```

### JSON Response
Return a JSON response from within a handler
```go
gomek.JSON(w, blog)
```

### CORS
Development CORS only
```go
app := gomek.New(gomek.Config{})
app.Use(gomek.CORS)
```

### Set BaseTemplates
Set the base templates via the `BaseTemplates` method
```go
app := gomek.New(gomek.Config{BaseTemplateName: "layout"})
app.BaseTemplates("./template/layout.html". "./templates/hero.html")
```
### Handlers with multiple methods
If you want to assign multiple verbs to the same route then use the following method clause
```go
app.Route("/blog").View(blog).Methods("GET", "POST")
func index(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
    if r.Method == "GET" {
		// code for GET requests to /blogs
    }
	if r. Method == "POST" {
		// code for POST requests to /blog
    }
	// .etc...
```

### Authorisation
If you use the `gomek.Authorize` middleware, all your routes will need to pass authorization
via the callback function passed to `gomek.Authorize`. To whitelist routes, pass a list of string
pairs, representing the path and the request method, respectively.
```go
var whiteList = [][]string{
		{
			"/", "GET",
		},
		{
			"/login", "GET",
		},
	}
```
The `gomek.Authorize` middleware function require 2 arguments, your `[][]string` of path / request methods
and a callback function to test your auth strategy (e.g. session  or JWT).
```go
app.Use(gomek.Authorize(whiteList, func(r *http.Request) (bool, context.Context) {
    // if your authorization test passes then return true
    return true, nil
}))
```
You can also attach values of any type to the request context
```go
app.Use(gomek.Authorize(whiteList, func(r *http.Request) (bool, context.Context) {
    // You can attach values to the request context & returning the context also
    ctx := context.WithValue(r.Context(), "userID", 1)
    return true, ctx
}))
```

### Restful approach
```go
// Create a type that represents your resource
type Notice struct {
}
// Implement the `Resource` interface
func (n *Notice) Post(w http.ResponseWriter, request *http.Request, d *gomek.Data) {
	panic("implement me")
}

func (n *Notice) Put(w http.ResponseWriter, request *http.Request, d *gomek.Data) {
	panic("implement me")
}

func (n *Notice) Delete(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
	panic("implement me")
}

func (n *Notice) Get(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
	var notice schemas.Notice
	notice.Name = "Joe!"
	gomek.JSON(w, notice, http.StatusOK)
}
```
To use your implementation of the `Resource` type
```go
// The Resource method expects a type that implements the `Resource` interface.
app.Route("/notices").Resource(&routes.Notice{}).Methods("GET")
```

### URL Path Variables
```go
app.Route("/blogs/<blog_id>").View(GetBlogs).Methods("GET", "POST")

func GetBlogs(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
    vars := gomek.Args(r)
    advertId := vars["blog_id"]
```

### Query Params
GetParams returns slices of string
```go
// example request url - http://127.0.0.1:8080/users?user_id=1
userID, err := gomek.GetParams(r, "user_id")
if err != nil {
	log.Println("no user_id in params")
	return
}
// userID[0] = "1"
```

### Set Host
Set the machine's host
```go
app.SetHost("127.0.0.1")
```


### Static Files
**Not specific to Gomek (example implements the standard library's static files setup)
```go
app := gomek.New(gomek.Config{})
publicFiles := http.FileServer(http.Dir("public"))
app.Handle("/public/", http.StripPrefix("/public/", publicFiles))
```

# Testing
Gomek provides a testing utility that returns a regular `HandlerFunc`
```go
c := Config{}
mockApp := NewTestApp(c)
mockApp.Route("/blogs").Methods("POST").Resource(&Notice{})
mockApp.Start()

notice := Notice{}

handler := CreateTestHandler(mockApp, notice.Post)

req := httptest.NewRequest(http.MethodPost, "/blogs", nil)
w := httptest.NewRecorder()
handler(w, req) // regular HandlerFunc
resp := w.Result()

defer resp.Body.Close()
data, err := io.ReadAll(resp.Body)
if err != nil {
t.Errorf("Error: %v", err)
}
expected := `{"name":"Joe"}`
if string(data) != expected {
t.Errorf("Expected %s got '%v'", expected, string(data))
```