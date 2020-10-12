package packaging

import "github.com/go-flutter-desktop/hover/internal/build"

type noopTask struct{}

var NoopTask Task = &noopTask{}

func (_ *noopTask) Name() string            { return "" }
func (_ *noopTask) Init()                   {}
func (_ *noopTask) IsInitialized() bool     { return true }
func (_ *noopTask) AssertInitialized()      {}
func (_ *noopTask) Pack(string, build.Mode) {}
func (_ *noopTask) AssertSupported()        {}
