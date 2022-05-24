package gomek

import "fmt"

// Template type used to hold the app's base templates
type Template struct {
	base []string
}

// Run creates a Template type & returns the current route's templates
func (t *Template) Run(routeTemplates ...string) []string {
	var currentViewTemplates []string
	currentViewTemplates = append(currentViewTemplates, t.base...)
	currentViewTemplates = append(currentViewTemplates, routeTemplates...)
	fmt.Println(currentViewTemplates)
	return currentViewTemplates
}
