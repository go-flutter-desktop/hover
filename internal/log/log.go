package log

import (
	"fmt"
	"github.com/logrusorgru/aurora"
)

// Au Aurora instance used for colors
var Au aurora.Aurora

// Print print a message with formatting
func Print(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(fmt.Sprintf(part, parts...))
}

// Fatal print a error with formatting (red)
func Fatal(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(Au.Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.RedFg).String())
}

// Warn print a warning with formatting (yellow)
func Warn(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(Au.Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.YellowFg).String())
}

// Info print a information with formatting (green)
func Info(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(Au.Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.GreenFg).String())
}

func hoverPrint() {
	fmt.Print(Au.Bold(Au.Cyan("hover: ")).String())
}
