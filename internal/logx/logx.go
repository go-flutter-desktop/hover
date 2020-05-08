package logx

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/logrusorgru/aurora"
)

type output interface {
	Output(depth int, s string) error
}

// Option for loggers.
type Option func(*Logger)

// OptionColorize enable colors for this logger.
func OptionColorize(l *Logger) {
	l.au = aurora.NewAurora(true)
}

var (
	defaultl = New()
	mu       sync.Mutex
)

// Tune the package level logger.
func Tune(options ...Option) {
	mu.Lock()
	defer mu.Unlock()
	defaultl = defaultl.Tune(options...)
}

// Printf print a message with formatting
func Printf(part string, parts ...interface{}) {
	defaultl.Output(2, fmt.Sprintf(part, parts...))
}

// Errorf print a error with formatting (red)
func Errorf(part string, parts ...interface{}) {
	defaultl.Output(2, defaultl.au.Colorize(fmt.Sprintf(part, parts...), aurora.RedFg).String())
}

// Warnf print a warning with formatting (yellow)
func Warnf(part string, parts ...interface{}) {
	defaultl.Output(2, defaultl.au.Colorize(fmt.Sprintf(part, parts...), aurora.YellowFg).String())
}

// Infof print a information with formatting (green)
func Infof(part string, parts ...interface{}) {
	defaultl.Output(2, defaultl.au.Colorize(fmt.Sprintf(part, parts...), aurora.GreenFg).String())
}

// Au default logger colorize.
func Au() aurora.Aurora {
	return defaultl.au
}

// New logger with colors disabled.
func New(options ...Option) Logger {
	return Logger{
		au:  aurora.NewAurora(false),
		out: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile),
	}.Tune(options...)
}

// Logger general logger for hover.
type Logger struct {
	au  aurora.Aurora
	out output
}

// Tune the logger with the provided options.
func (t Logger) Tune(options ...Option) Logger {
	for _, opt := range options {
		opt(&t)
	}

	return t
}

// Au return the reference to colorize
func (t Logger) Au() aurora.Aurora {
	return t.au
}

// Output see log.Output
func (t Logger) Output(depth int, s string) error {
	return t.out.Output(depth+1, t.prefix()+s)
}

func (t Logger) prefix() string {
	return t.au.Colorize("hover: ", aurora.CyanFg).String()
}
