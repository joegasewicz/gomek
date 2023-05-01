package gomek

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Notice struct{}

func (n *Notice) Post(w http.ResponseWriter, request *http.Request, d *Data) {
	data := struct {
		Name string `json:"name"`
	}{
		Name: "Joe",
	}
	JSON(w, data, http.StatusOK)
}

func (n *Notice) Put(w http.ResponseWriter, request *http.Request, d *Data) {
	data := struct {
		Name string `json:"name"`
	}{
		Name: "Cosmo",
	}
	JSON(w, data, http.StatusOK)
}

func (n *Notice) Delete(w http.ResponseWriter, r *http.Request, d *Data) {
	data := struct {
		Name string `json:"name"`
	}{
		Name: "Alex",
	}
	JSON(w, data, http.StatusOK)
}

func (n *Notice) Get(w http.ResponseWriter, r *http.Request, d *Data) {
	data := struct {
		Name string `json:"name"`
	}{
		Name: "Ram",
	}
	JSON(w, data, http.StatusOK)
}

func TestViewGet(t *testing.T) {
	c := Config{}
	mockApp := NewTestApp(c)
	mockApp.Route("/blogs").Methods("POST").Resource(&Notice{})
	mockApp.Start()

	notice := Notice{}

	handler := CreateTestHandler(mockApp, notice.Get)

	req := httptest.NewRequest(http.MethodPost, "/blogs", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	resp := w.Result()

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	expected := `{"name":"Joe"}`
	if string(data) != expected {
		t.Errorf("Expected %s got '%v'", expected, string(data))
	}
}

func TestViewPost(t *testing.T) {
	c := Config{}
	mockApp := NewTestApp(c)
	mockApp.Route("/blogs").Methods("GET", "POST").Resource(&Notice{})
	mockApp.Start()

	notice := Notice{}

	handler := CreateTestHandler(mockApp, notice.Get)

	req := httptest.NewRequest(http.MethodGet, "/blogs", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	resp := w.Result()

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	expected := `{"name":"Ram"}`
	if string(data) != expected {
		t.Errorf("Expected %s got '%v'", expected, string(data))
	}
}

func TestViewDelete(t *testing.T) {
	c := Config{}
	mockApp := NewTestApp(c)
	mockApp.Route("/blogs").Methods("GET", "POST", "PUT", "DELETE").Resource(&Notice{})
	mockApp.Start()

	notice := Notice{}

	handler := CreateTestHandler(mockApp, notice.Put)

	req := httptest.NewRequest(http.MethodPut, "/blogs", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	resp := w.Result()

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	expected := `{"name":"Cosmo"}`
	if string(data) != expected {
		t.Errorf("Expected %s got '%v'", expected, string(data))
	}
}
