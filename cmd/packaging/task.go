package packaging

// Task contains all configuration options for a given packaging method.
// TODO: Rename to something that suits it more? Mabe Executor?
type Task interface {
	Name() string
	Init()
	IsInitialized() bool
	AssertInitialized()
	Pack(buildVersion string)
}
