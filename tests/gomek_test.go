package tests

import (
	"github.com/joegasewicz/gomek"
	"net/http"
	"testing"
	"time"
)

type JsonTest struct {
	Name string `json:"name"`
}

func GetHandler(w http.ResponseWriter, r *http.Request, d *gomek.Data) {
	var j JsonTest
	j.Name = "Joe"
	gomek.JSON(w, j, http.StatusOK)
}

func TestStart(t *testing.T) {
	c := gomek.Config{
		BaseTemplateName: "base.gohtml",
		BaseTemplates: []string{
			"template1.gohtml",
			"template2.gohtml",
			"template3.gohtml",
		},
	}
	app := gomek.New(c)
	app.Route("/").View(GetHandler).Methods("GET")
	app.Listen(2001)
	app.Host("localhost")

	chan1 := make(chan struct{})
	chan2 := make(chan gomek.App)

	go func() {
		close(chan1)
		app.Start()
		chan2 <- app
		close(chan2)
	}()
	<-chan1
	time.Sleep(1 * time.Second)
	if app.Config.BaseTemplateName != "base.gohtml" {
		t.Error("expected base template to be base.gohtml but got ", app.Config.BaseTemplateName)
	}
	if app.CurrentHost != "localhost" {
		t.Error("expected CurrentHost to be localhost but got", app.CurrentHost)
	}
	if app.Port != 2001 {
		t.Error("expected Port to be 2001 but got", app.Port)
	}
	if app.Protocol != "http" {
		t.Error("Expected protocol to be http but got ", app.Protocol)
	}
	if len(app.CurrentMethods) != 1 {
		t.Error("expected 1 handler to be added but got", len(app.CurrentMethods))
	}

	app.Shutdown()
	<-chan2
}
