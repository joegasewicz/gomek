package gomek

import "net/http"

// CreateTestHandler this function provides an easy way to access the
// `http.HandlerFunc` func. Testing gomek views is thus the same as testing
// any regular HandlerFunc handlers.
//
//	c := Config{}
//	mockApp := NewTestApp(c)
//	mockApp.Route("/blogs").Methods("POST").Resource(&Notice{})
//	mockApp.Start()
//
//	notice := Notice{}
//
//	handler := CreateTestHandler(mockApp, notice.Post)
//
//	req := httptest.NewRequest(http.MethodPost, "/blogs", nil)
//	w := httptest.NewRecorder()
//	handler(w, req) // regular HandlerFunc
//	resp := w.Result()
//
//	defer resp.Body.Close()
//	data, err := io.ReadAll(resp.Body)
//	if err != nil {
//	t.Errorf("Error: %v", err)
//	}
//	expected := `{"name":"Joe"}`
//	if string(data) != expected {
//	t.Errorf("Expected %s got '%v'", expected, string(data))
func CreateTestHandler(testApp IApp, view CurrentView) http.HandlerFunc {
	v := testApp.GetView()
	var templates []string
	return v.handleFuncWrapper(templates, testApp.GetConfig(), *testApp.GetView(), view)
}
