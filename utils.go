package gomek

import "fmt"

const (
	RESET = "\033[0m"
	RED   = "\033[31m"
	GREEN = "\033[32m"
	BLUE  = "\033[34m"
)

func PrintWithColor(msg string, color string) string {
	return fmt.Sprintf("%s%s%s", color, msg, RESET)
}
