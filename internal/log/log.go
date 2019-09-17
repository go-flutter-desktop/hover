package log

import (
	"fmt"
	"github.com/logrusorgru/aurora"
)

var Au aurora.Aurora

func Print(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(fmt.Sprintf(part, parts...))
}

func Fatal(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(Au.Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.RedFg).String())
}

func Warn(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(Au.Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.YellowFg).String())
}

func Info(part string, parts ...interface{}) {
	hoverPrint()
	fmt.Println(Au.Colorize(fmt.Sprintf(fmt.Sprintf("%v", part), parts...), aurora.GreenFg).String())
}

func hoverPrint() {
	fmt.Print(Au.Colorize("hover:", aurora.CyanBg|aurora.BlackFg).String() + " ")
}
