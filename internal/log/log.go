package log

import (
	"fmt"
	"log"

	"github.com/logrusorgru/aurora"
)

func init() {
	EnableColoredConsoleOutput()
	log.SetFlags(0)
}

// internal colorized
var au aurora.Aurora = aurora.NewAurora(false)

// internal verbosity level
var verbose bool

// Verbosity enable verbose logging
func Verbosity(b bool) {
	verbose = b
	if b {
		log.SetFlags(log.Flags() | log.Lshortfile | log.Ltime)
	} else {
		log.SetFlags(0)
	}
}

// Colorize set the logger to support colors printing.
func Colorize(b bool) {
	au = aurora.NewAurora(b)
}

// Au Aurora instance used for colors
func Au() aurora.Aurora {
	return au
}

// Printf print a message with formatting
func Printf(part string, parts ...interface{}) {
	log.Output(2, fmt.Sprint(
		hoverPrint(),
		fmt.Sprintf(part, parts...),
	))
}

// Errorf print a error with formatting (red)
func Errorf(part string, parts ...interface{}) {
	log.Output(2, fmt.Sprint(
		hoverPrint(),
		Au().Colorize(fmt.Sprintf(part, parts...), aurora.RedFg).String(),
	))
}

// Warnf print a warning with formatting (yellow)
func Warnf(part string, parts ...interface{}) {
	log.Output(2, fmt.Sprint(
		hoverPrint(),
		Au().Colorize(fmt.Sprintf(part, parts...), aurora.YellowFg).String(),
	))
}

// Infof print a information with formatting (green)
func Infof(part string, parts ...interface{}) {
	log.Output(2, fmt.Sprint(
		hoverPrint(),
		Au().Colorize(fmt.Sprintf(part, parts...), aurora.GreenFg).String(),
	))
}

// Debugf print debugging information with formatting (green)
func Debugf(part string, parts ...interface{}) {
	if !verbose {
		return
	}

	log.Output(2, fmt.Sprint(
		hoverPrint(),
		Au().Faint(fmt.Sprintf(part, parts...)).String(),
	))
}

func hoverPrint() string {
	return Au().Bold(Au().Cyan("hover: ")).String()
}
