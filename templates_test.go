package gomek

import "testing"

func TestTemplate_Run(t *testing.T) {
	template := Template{base: []string{"base.html"}}
	routeTemplates := []string{"/index.gohtml", "/news.gohtml"}
	result := template.Run(routeTemplates...)
	expected := []string{"base.html", "/index.gohtml", "/news.gohtml"}
	for i, _ := range result {
		if result[i] != expected[i] {
			t.Errorf("expected %v got %v", result[i], expected[i])
		}
	}

}
