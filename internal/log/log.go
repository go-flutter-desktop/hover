package log

import (
	"fmt"

	"github.com/logrusorgru/aurora"
)

// Colorize chnage the logger to support colors printing.
func Colorize() {
	au = aurora.NewAurora(true)
}

// internal colorized
var au aurora.Aurora

// Au Aurora instance used for colors
func Au() aurora.Aurora {
	if au == nil {
		au = aurora.NewAurora(false)
	}
	return au
}

// Printf print a message with formatting
func Printf(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(fmt.Sprintf(part, parts...))
}

// Errorf print a error with formatting (red)
func Errorf(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(Au().Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.RedFg).String())
}

// Warnf print a warning with formatting (yellow)
func Warnf(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(Au().Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.YellowFg).String())
}

// Infof print a information with formatting (green)
func Infof(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(Au().Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.GreenFg).String())
}

func hoverPrint() {
	fmt.Print(Au().Bold(Au().Cyan("hover: ")).String())
}
