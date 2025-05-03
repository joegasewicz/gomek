package gomek

import "fmt"

// Template type used to hold the app's base registeredTemplates
type Template struct {
	base           []string
	RouteTemplates []string
}

// Run creates a Template type & returns the current route's registeredTemplates
func (t *Template) Run(routeTemplates ...string) []string {
	var currentViewTemplates []string
	currentViewTemplates = append(currentViewTemplates, t.base...)
	currentViewTemplates = append(currentViewTemplates, routeTemplates...)
	return currentViewTemplates
}

func LogTemplates(registeredTemplates []RegisteredTemplates) {
	if len(registeredTemplates) > 0 {
		out := PrintWithColor("[Registering Templates]:", BLUE)
		fmt.Println(out)
		for _, registeredTemplate := range registeredTemplates {
			msg := fmt.Sprintf("\t- Route: %s\n", registeredTemplate.Route)
			out = PrintWithColor(msg, GREEN)
			fmt.Printf(out)
			msg = fmt.Sprintf("\t- Route Templates: %v\n", registeredTemplate.Templates)
			out = PrintWithColor(msg, GREEN)
			fmt.Printf(out)
			msg = fmt.Sprintf("\t- Partial Templates: \n")
			out = PrintWithColor(msg, GREEN)
			fmt.Printf(out)
			for _, partial := range registeredTemplate.Partials {
				msg = fmt.Sprintf("\t\t-  %s\n", partial)
				out = PrintWithColor(msg, GREEN)
				fmt.Printf(out)
			}
			fmt.Printf("\n")
		}
	}
}
