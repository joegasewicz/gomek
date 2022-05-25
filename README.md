# Gomek
A tiny http framework

## Features
- Easy to learn API.
- Templates are managed for you - only declare your base templates once.
- Views handle your data only e.g. no template logic (Single-responsibility principle).
- Define HTTP methods per View method.
- Middleware - add yor own middleware
- Logging
- CORS
- JSON

# Example
```go
import (
    "github.com/joegasewicz/gomek"
    "net/http"
)
// A view that handles a template's data
func index(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
    templateData := make(gomek.Data)
	// templateData["Title"] = "Hello!"
	*d = templateData
}

// A view that just returns JSON
func blog(w http.ResponseWriter, r *http.Request, data *gomek.Data) {
    var blog Blog
	// query database...
	gomek.JSON(w, blog)
}

func main() {
    c := gomek.Config{
        // Declare your base template e.g (`layout` if template filename is `layout.gohtml`)
        BaseTemplateName: "layout",
		BaseTemplates: []string{
            "./templates/layout.gohtml",
            "./templates/sidebar.gohtml",
            "./templates/navbar.gohtml",
            "./templates/footer.gohtml",
        },
		// Or declare your base template paths with `BaseTemplates([]string)`
		// app.BaseTemplates("./templates/layout.gohtml")
    }
	// Create a new gomek app
    app := gomek.New(c)
	// Handle static files
	files := http.FileServer(http.Dir("static"))
    app.Handle("/static/", http.StripPrefix("/static/", files))
    // Declare your views
    app.Route("/blog").View(blog).Methods("GET", "POST").Templates("./templates/blog.gohtml")
    app.Route("/").View(index).Methods("GET").Templates("./templates/index.gohtml")
	// Add middleware
    app.Use(gomek.CORS) // Use gomek's CORS or any other third party package
	app.Use(MyCustomAuth)
    // Set a port (optional)
    app.Listen(6011)
	// Start your app
    app.Start()
}
```