// Package debugx controls debug logging.
package debugx

import (
	"fmt"
	"sync"

	"github.com/go-flutter-desktop/hover/internal/logx"
)

var (
	defaultl = logx.New(logx.OptionDiscard)
	mu       sync.Mutex
)

// Tune the package level logger.
func Tune(options ...logx.Option) {
	mu.Lock()
	defer mu.Unlock()
	defaultl = defaultl.Tune(options...)
}

// Println see fmt.Sprintln
func Println(parts ...interface{}) {
	defaultl.Output(2, fmt.Sprintln(parts...))
}

// Printf print a message with formatting
func Printf(part string, parts ...interface{}) {
	defaultl.Output(2, fmt.Sprintf(part, parts...))
}
