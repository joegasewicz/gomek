# Gomek
A tiny http framework

## Features
- Easy to learn API.
- Templates are managed for you - only declare your base templates once.
- Views handle your data only e.g. no template logic (Single-responsibility principle).
- Define HTTP methods per View method.


# Example
```go
import (
    "github.com/joegasewicz/gomek"
    "net/http"
)
// Create your views
func index(w http.ResponseWriter, r *http.Request, data *gomek.Data) {
    *data = nil
}
func blog(w http.ResponseWriter, r *http.Request, data *gomek.Data) {
    *data = nil
}

func main() {
    c := gomek.Config{
        // Declare your base template e.g (`layout` if template filename is `layout.gohtml`)
        BaseTemplateName: "layout", 
    }
	// Create a new gomek app
    app := gomek.New(c)
	// Handle static files
	files := http.FileServer(http.Dir("static"))
    app.Handle("/static/", http.StripPrefix("/static/", files))
    // Declare your views
    app.Route("/blog").View(blog).Methods("GET", "POST").Templates("./templates/blog.gohtml")
    app.Route("/").View(index).Methods("GET").Templates("./templates/index.gohtml")
    // Declare all your base template paths
    app.BaseTemplates("./templates/layout.gohtml")
    app.Listen(6011) // Set a port (optional)
	// Start your app
    app.Start()
}
```